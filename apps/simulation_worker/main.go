// Package main pulls simulation requests from the
// simulation queue, executes the simulations using simc, and
// then writes the results back to the database.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"time"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rabbitmq/amqp091-go"

	api_types "github.com/DomNidy/saint_sim/pkg/go-shared/api_types"
	dbqueries "github.com/DomNidy/saint_sim/pkg/go-shared/db"
	secrets "github.com/DomNidy/saint_sim/pkg/go-shared/secrets"
	utils "github.com/DomNidy/saint_sim/pkg/go-shared/utils"
)

var ErrInvalidSimInput = errors.New("invalid sim input")

var simcBinaryPath = secrets.LoadSecret("SIMC_BINARY_PATH").Value()

var queue *utils.SimulationQueueClient

func init() {
	user := secrets.LoadSecret("RABBITMQ_USER").Value()
	pass := secrets.LoadSecret("RABBITMQ_PASS").Value()
	host := secrets.LoadSecret("RABBITMQ_HOST").Value()
	port := secrets.LoadSecret("RABBITMQ_PORT").Value()

	simQueueClient, err := utils.NewSimulationQueueClient("saint_api", user, pass, host, port)
	if err != nil {
		log.Panicf("ERROR: Failed to initialize connection to simulation queue: %v", err)

		return
	}

	queue = simQueueClient
}

func validateSimInput(region, realm, name string) error {
	validPart := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	if region == "" || realm == "" || name == "" {
		return fmt.Errorf("%w: region, realm, and name must be non-empty", ErrInvalidSimInput)
	}

	if !validPart.MatchString(region) {
		return fmt.Errorf("%w: invalid region %q", ErrInvalidSimInput, region)
	}

	if !validPart.MatchString(realm) {
		return fmt.Errorf("%w: invalid realm %q", ErrInvalidSimInput, realm)
	}

	if !validPart.MatchString(name) {
		return fmt.Errorf("%w: invalid name %q", ErrInvalidSimInput, name)
	}

	return nil
}

func performSim(region, realm, name string) (*[]byte, error) {
	err := validateSimInput(region, realm, name)
	if err != nil {
		return nil, err
	}

	// Command to invoke simc and perform the sim
	// #nosec G204 - inputs are validated first and passed w/ a shell
	simCommand := exec.CommandContext(
		context.TODO(),
		simcBinaryPath,
		fmt.Sprintf("armory=%s,%s,%s", region, realm, name),
	)

	// Capture output of sim command and write it to this buffer
	var outputBuffer bytes.Buffer

	simCommand.Stdout = &outputBuffer
	simCommand.Stderr = os.Stderr

	// Run the sim command
	err = simCommand.Run()
	if err != nil {
		return nil, fmt.Errorf("%w: Error executing simc binary", err)
	}

	// Get the output as a byte slice
	simResult := outputBuffer.Bytes()

	return &simResult, nil
}

func getSimcVersion() string {
	simcCommand := exec.CommandContext(context.TODO(), simcBinaryPath)

	var outputBuffer bytes.Buffer

	simcCommand.Stdout = &outputBuffer

	err := simcCommand.Run()

	exitCode := simcCommand.ProcessState.ExitCode()

	const noArgumentsExitCode = 50 // simc returns exitcode 50 when no arguments are provided.
	// since we just want to read its stdout to parse version number, we can ignore the err

	if err != nil && exitCode != noArgumentsExitCode {
		log.Fatalf("Error running simc binary: %v", err)
	}

	res := outputBuffer.String()

	return res
}

func processSimulationMessage(ctx context.Context, msg amqp091.Delivery, dbClient dbqueries.Queries) {
	var simRequestMsg utils.SimulationMessage

	err := json.Unmarshal(msg.Body, &simRequestMsg)
	if err != nil {
		log.Printf(
			"WARNING: Simulation worker received message that could not be unmarshalled to json: %v",
			err,
		)

		return
	}

	// Query the sim options json object from simulation_request table
	var requestID pgtype.UUID

	err = requestID.Scan(simRequestMsg.SimulationID)
	if err != nil {
		log.Printf("Error converting simulation request id to uuid: %v", err)

		return
	}

	simOptionsJSON, err := dbClient.GetSimulationOptions(ctx, requestID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf(
				"Failed to locate simulation request to resolve options. Cannot process request: %v",
				err,
			)

			return
		}

		log.Printf(
			"Error occurred while trying to resolve simulation request options: %v",
			err,
		)

		return
	}

	// Validating the returned json from db conforms to the API shape
	var simOptions api_types.SimulationOptions

	err = json.Unmarshal(simOptionsJSON, &simOptions)
	if err != nil {
		log.Printf("error unmarshalling json: %v", err)

		return
	}

	log.Print("Received simulation request with options:")
	log.Printf("  %v", string(simOptionsJSON))

	if !utils.IsValidSimOptions(&simOptions) {
		log.Printf(
			"Invalid sim options received, potential RCE attempted: %v",
			string(simOptionsJSON),
		)

		return
	}

	startedAt := pgtype.Timestamptz{
		Time:             time.Now(),
		Valid:            true,
		InfinityModifier: pgtype.Finite,
	}

	_, err = dbClient.UpdateSimulation(
		context.Background(),
		dbqueries.UpdateSimulationParams{
			SimResult: pgtype.Text{String: "", Valid: false},
			StartedAt: startedAt,
			CompletedAt: pgtype.Timestamptz{
				Time:             time.Time{},
				Valid:            false,
				InfinityModifier: pgtype.Finite,
			},
			ID:        requestID,
			ErrorText: pgtype.Text{String: "", Valid: false},
		},
	)
	if err != nil {
		log.Printf(
			"WARNING: Failed to update the started at time for a simulation request. We will still process it but since" +
				"we failed to write the started at time, we may fail to write the results as well.",
		)
	}

	simulationResult, err := performSim(
		string(simOptions.WowCharacter.Region),
		string(simOptions.WowCharacter.Realm),
		simOptions.WowCharacter.CharacterName,
	)
	if err != nil {
		log.Printf("error while performing sim: %v", err)

		return
	}

	var simResPg pgtype.Text

	err = simResPg.Scan(string(*simulationResult))
	if err != nil {
		log.Printf("%v", err)

		return
	}

	var completedAt pgtype.Timestamptz

	err = completedAt.Scan(time.Now())
	if err != nil {
		log.Printf("error converting completed_at timestamp: %v", err)

		return
	}

	_, err = dbClient.UpdateSimulation(
		context.Background(),
		dbqueries.UpdateSimulationParams{
			SimResult:   simResPg,
			CompletedAt: completedAt,
			ID:          requestID,
			StartedAt: pgtype.Timestamptz{
				Time:             time.Time{},
				Valid:            false,
				InfinityModifier: 0,
			},
			ErrorText: pgtype.Text{String: "", Valid: false},
		},
	)
	if err != nil {
		log.Printf("error trying to insert sim data to db: %v", err)

		return
	}
}

func main() {
	log.Printf("simulation_worker, running SimC version: %s", getSimcVersion())

	ctx := context.Background()

	pool := utils.InitPostgresConnectionPool(ctx)
	dbClient := dbqueries.New(pool)

	defer pool.Close()
	defer queue.Close()

	// Immediately start receiving queued messages
	msgChan, err := queue.Consume(
		"",    // consumer
		true,  // auto ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	utils.FailOnError(err, "Failed to register as consumer")

	var receivedCount uint64

	go func() {
		for msg := range msgChan {
			receivedCount++

			log.Printf("Received a message: %s\n", msg.Body)
			log.Printf("receivedCount = %d\n", receivedCount)
			processSimulationMessage(ctx, msg, *dbClient)
		}
	}()

	select {}
}
