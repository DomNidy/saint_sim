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

	"github.com/DomNidy/saint_sim/apps/simulation_worker/data"
	interfaces "github.com/DomNidy/saint_sim/pkg/interfaces"
	secrets "github.com/DomNidy/saint_sim/pkg/secrets"
	utils "github.com/DomNidy/saint_sim/pkg/utils"
	"github.com/jackc/pgx/v5"
)

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
	log.Printf("SIMC Version: %s", getSimcVersion())

	ctx := context.Background()
	// Setup postgres connection
	db := utils.InitPostgresConnectionPool(ctx)
	defer db.Close()

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
	var receivedCount uint64

	go func() {
		for d := range msgs {
			receivedCount += 1
			log.Printf("Received a message: %s\n", d.Body)
			log.Printf("receivedCount = %d\n", receivedCount)

			var simRequestMsg interfaces.SimulationMessageBody

			// todo: handle this error, and finish this message
			err := json.Unmarshal(d.Body, &simRequestMsg)
			if err != nil {
				log.Printf("error unmarshalling json: %v", err)
				continue
			}

			// Query the sim options json object from simulation_request table
			var simOptionsJson []byte
			err = db.QueryRow(context.Background(), "select options from simulation_request where id = $1", simRequestMsg.SimulationId).Scan(&simOptionsJson)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					log.Printf("Failed to locate simulation request to resolve options. Cannot process request: %v", err)
					continue
				}
				log.Printf("Error occured while trying to resolve simulation request options: %v", err)
				continue
			}

			// Validating the returned json from db
			var simOptions interfaces.SimulationOptions
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
			err = data.InsertSimulationData(db, &interfaces.SimDataInsert{
				RequestID: *simRequestMsg.SimulationId,
				SimResult: string(*simulationResult),
			})
			if err != nil {
				log.Printf("error trying to insert sim data to db: %v", err)
				continue
			}
		}
	}()
	<-forever
}
