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

	api_types "github.com/DomNidy/saint_sim/pkg/go-shared/api_types"
	dbqueries "github.com/DomNidy/saint_sim/pkg/go-shared/db"
	secrets "github.com/DomNidy/saint_sim/pkg/go-shared/secrets"
	utils "github.com/DomNidy/saint_sim/pkg/go-shared/utils"
	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimulationStore interface {
	GetSimulationRequestOptions(ctx context.Context, id pgtype.UUID) ([]byte, error)
	InsertSimulationData(ctx context.Context, arg dbqueries.InsertSimulationDataParams) error
}

type SimRunner interface {
	Perform(options string) ([]byte, error)
	Version() string
}

type Worker struct {
	store  SimulationStore
	runner SimRunner
}

type simcRunner struct {
	binaryPath string
}

func (r simcRunner) Perform(options string) ([]byte, error) {
	// Command to invoke simc and perform the sim
	simCommand := exec.Command(r.binaryPath, options)

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

	return simResult, nil
}

func (r simcRunner) Version() string {
	log.Print(r.binaryPath)
	simcCommand := exec.Command(r.binaryPath)

	var outputBuffer bytes.Buffer
	simcCommand.Stdout = &outputBuffer

	err := simcCommand.Run()

	exitCode := simcCommand.ProcessState.ExitCode()
	const noArgumentsExitCode = 50 // simc returns exitcode 50 when no arguments are provided.
	// since we just want to read its stdout to parse version number, we can ignore the err

	if err != nil && exitCode != noArgumentsExitCode {
		log.Fatalf("Error running simc binary: %v", err)
	}

	return outputBuffer.String()
}

func (w Worker) HandleMessage(ctx context.Context, body []byte) error {
	var simRequestMsg api_types.SimulationMessageBody
	if err := json.Unmarshal(body, &simRequestMsg); err != nil {
		return fmt.Errorf("unmarshal simulation message: %w", err)
	}

	if simRequestMsg.SimulationId == nil {
		return errors.New("simulation_id is required")
	}

	var requestID pgtype.UUID
	if err := requestID.Scan(*simRequestMsg.SimulationId); err != nil {
		return fmt.Errorf("convert simulation request id to uuid: %w", err)
	}

	simOptionsJSON, err := w.store.GetSimulationRequestOptions(ctx, requestID)
	if err != nil {
		return fmt.Errorf("resolve simulation request options: %w", err)
	}

	var simOptions api_types.SimulationOptions
	if err := json.Unmarshal(simOptionsJSON, &simOptions); err != nil {
		return fmt.Errorf("unmarshal simulation options: %w", err)
	}

	log.Print("Received simulation request with options:")
	log.Printf("  %v", string(simOptionsJSON))

	if !utils.IsValidSimOptions(&simOptions) {
		return fmt.Errorf("invalid sim options received: %s", string(simOptionsJSON))
	}

	region := string(simOptions.WowCharacter.Region)
	realm := string(simOptions.WowCharacter.Realm)
	characterName := simOptions.WowCharacter.CharacterName
	simulationResult, err := w.runner.Perform(fmt.Sprintf("armory=%v,%v,%v", region, realm, characterName))

	if err != nil {
		return fmt.Errorf("perform sim: %w", err)
	}

	if err := w.store.InsertSimulationData(ctx, dbqueries.InsertSimulationDataParams{
		RequestID: requestID,
		SimResult: string(simulationResult),
	}); err != nil {
		return fmt.Errorf("insert simulation data: %w", err)
	}

	return nil
}

func consumeMessages(ctx context.Context, worker Worker, msgs <-chan amqp.Delivery) {
	var receivedCount uint64

	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-msgs:
			if !ok {
				return
			}

			receivedCount += 1
			log.Printf("Received a message: %s\n", d.Body)
			log.Printf("receivedCount = %d\n", receivedCount)

			err := worker.HandleMessage(ctx, d.Body)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					log.Printf("Failed to locate simulation request to resolve options. Cannot process request: %v", err)
					continue
				}
				log.Printf("Failed to process simulation request: %v", err)
			}
		}
	}
}

func main() {
	binaryPath := secrets.LoadSecret("SIMC_BINARY_PATH").Value()
	runner := simcRunner{binaryPath: binaryPath}
	log.Printf("SIMC Version: %s", runner.Version())

	ctx := context.Background()

	// Setup postgres connection
	db := utils.InitPostgresConnectionPool(ctx)
	defer db.Close()

	worker := Worker{
		store:  dbqueries.New(db),
		runner: runner,
	}

	// Setup rabbit mq connection
	conn, ch := utils.InitRabbitMQConnection()
	defer conn.Close()
	defer ch.Close()

	// declare queue
	q := utils.DeclareSimulationQueue(ch)

	// Immediately start receiving queued messages
	msgs, err := ch.Consume(
		q.Name,
		"",    // consumer
		true,  // auto ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	utils.FailOnError(err, "Failed to register as consumer")

	var forever chan struct{}

	go consumeMessages(ctx, worker, msgs)
	<-forever
}
