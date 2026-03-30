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
	"time"

	api_types "github.com/DomNidy/saint_sim/pkg/go-shared/api_types"
	dbqueries "github.com/DomNidy/saint_sim/pkg/go-shared/db"
	secrets "github.com/DomNidy/saint_sim/pkg/go-shared/secrets"
	utils "github.com/DomNidy/saint_sim/pkg/go-shared/utils"
	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var queue *utils.SimulationQueueClient = nil

func init() {
	user := secrets.LoadSecret("RABBITMQ_USER").Value()
	pass := secrets.LoadSecret("RABBITMQ_PASS").Value()
	host := secrets.LoadSecret("RABBITMQ_HOST").Value()
	port := secrets.LoadSecret("RABBITMQ_PORT").Value()
	q, err := utils.NewSimulationQueueClient("saint_api", user, pass, host, port)
	if err != nil {
		log.Panicf("ERROR: Failed to initialize connection to simulation queue: %v", err)
		return
	}
	queue = q
}

var SIMC_BINARY_PATH = secrets.LoadSecret("SIMC_BINARY_PATH").Value()

func performSim(region, realm, name string) (*[]byte, error) {
	// Command to invoke simc and perform the sim
	simCommand := exec.Command(SIMC_BINARY_PATH, fmt.Sprintf("armory=%v,%v,%v", region, realm, name))

	// Capture output of sim command and write it to this buffer
	var outputBuffer bytes.Buffer
	simCommand.Stdout = &outputBuffer
	simCommand.Stderr = os.Stderr

	// Run the sim command
	if err := simCommand.Run(); err != nil {
		log.Printf("Failed to execute sim binary: %v", err)
		return nil, err
	}

	// Get the output as a byte slice
	simResult := outputBuffer.Bytes()

	return &simResult, nil
}

func getSimcVersion() string {
	log.Print(SIMC_BINARY_PATH)
	simcCommand := exec.Command(SIMC_BINARY_PATH)

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

func main() {
	log.Printf("simulation_worker, running SimC version: %s", getSimcVersion())
	ctx := context.Background()

	db := utils.InitPostgresConnectionPool(ctx)
	queries := dbqueries.New(db)

	defer db.Close()
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
		for d := range msgChan {
			receivedCount += 1
			log.Printf("Received a message: %s\n", d.Body)
			log.Printf("receivedCount = %d\n", receivedCount)

			var simRequestMsg utils.SimulationMessage

			// todo: handle this error, and finish this message
			err := json.Unmarshal(d.Body, &simRequestMsg)
			if err != nil {
				log.Printf("error unmarshalling json: %v", err)
				continue
			}

			// Query the sim options json object from simulation_request table
			var requestID pgtype.UUID
			err = requestID.Scan(simRequestMsg.SimulationID)
			if err != nil {
				log.Printf("Error converting simulation request id to uuid: %v", err)
				continue
			}

			simOptionsJson, err := queries.GetSimulationOptions(context.Background(), requestID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					log.Printf("Failed to locate simulation request to resolve options. Cannot process request: %v", err)
					continue
				}
				log.Printf("Error occured while trying to resolve simulation request options: %v", err)
				continue
			}

			// Validating the returned json from db conforms to the API shape
			var simOptions api_types.SimulationOptions
			err = json.Unmarshal(simOptionsJson, &simOptions)
			if err != nil {
				log.Printf("error unmarshalling json: %v", err)
				continue
			}

			log.Print("Received simulation request with options:")
			log.Printf("  %v", string(simOptionsJson))

			if !utils.IsValidSimOptions(&simOptions) {
				log.Printf("Invalid sim options received, potential RCE attempted: %v", string(simOptionsJson))
				continue
			}

			simulationResult, err := performSim(string(simOptions.WowCharacter.Region), string(simOptions.WowCharacter.Realm), simOptions.WowCharacter.CharacterName)
			if err != nil {
				log.Printf("error while performing sim: %v", err)
				continue
			}

			var simResPg pgtype.Text
			if err = simResPg.Scan(string(*simulationResult)); err != nil {
				log.Printf("%v", err)
				continue
			}
			var completedAt pgtype.Timestamptz
			if err = completedAt.Scan(time.Now()); err != nil {
				log.Printf("error converting completed_at timestamp: %v", err)
				continue
			}
			_, err = queries.UpdateSimulation(context.Background(), dbqueries.UpdateSimulationParams{
				SimResult:   simResPg,
				CompletedAt: completedAt,
				ID:          requestID,
			})
			if err != nil {
				log.Printf("error trying to insert sim data to db: %v", err)
				continue
			}
		}
	}()
	select {}
}
