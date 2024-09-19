package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	interfaces "github.com/DomNidy/saint_sim/pkg/interfaces"
	secrets "github.com/DomNidy/saint_sim/pkg/secrets"
	utils "github.com/DomNidy/saint_sim/pkg/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

func performSim(region, realm, name string) {

	simcBinaryPath := secrets.LoadSecret("SIMC_BINARY_PATH")
	outputFilePath := fmt.Sprintf("%v-%v-%v-%d.txt", region, realm, name, time.Now().Unix())
	simCommand := exec.Cmd{
		Path:   simcBinaryPath.Value(),
		Args:   []string{simcBinaryPath.Value(), fmt.Sprintf("armory=%v,%v,%v", region, realm, name), fmt.Sprintf("output=%v", outputFilePath)},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	fmt.Println(simCommand.Args)
	if err := simCommand.Run(); err != nil {
		log.Fatalf("Failed to execute sim binary: %v", err)
	}

	data, err := os.ReadFile(outputFilePath)
	if err != nil {
		log.Fatalf("Failed to read sim output file: %v", err)
	}
	fmt.Printf("Sim data: %v\n", data)
}

func main() {
	// Setup rabitmq
	RABBITMQ_USER := secrets.LoadSecret("RABBITMQ_USER")
	RABBITMQ_PASS := secrets.LoadSecret("RABBITMQ_PASS")
	RABBITMQ_PORT := secrets.LoadSecret("RABBITMQ_PORT")
	RABBITMQ_HOST := secrets.LoadSecret("RABBITMQ_HOST")
	connectionURI := fmt.Sprintf("amqp://%s:%s@%s:%s", RABBITMQ_USER.Value(), RABBITMQ_PASS.Value(), RABBITMQ_HOST.Value(), RABBITMQ_PORT.Value())
	// Connect to rabbitmq
	conn, err := amqp.Dial(connectionURI)
	utils.FailOnError(err, "Failed to connect to rabbitmq")
	defer conn.Close()

	// setup channel
	ch, err := conn.Channel()
	utils.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	// declare queue
	q, err := ch.QueueDeclare(
		"simulation_queue", // name
		false,              // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	utils.FailOnError(err, "Failed to declare a queue")

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
	utils.FailOnError(err, "Failed to register a consumer")

	var forever chan struct{}
	var receivedCount uint64

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s\n", d.Body)
			log.Printf("receivedCount = %d\n", receivedCount)

			var simRequestMsg interfaces.SimulateJSONRequestBody

			// todo: figure out how to validate the request object
			err := json.Unmarshal(d.Body, &simRequestMsg)
			if err != nil {
				log.Printf("error unmarshalling json: %v", err)
			}
			fmt.Println("Received simulation request:")
			fmt.Printf("	character_name: %s\n", *simRequestMsg.WowCharacter.CharacterName)
			fmt.Printf("	realm: %s\n", *simRequestMsg.WowCharacter.Realm)
			fmt.Printf("	region: %s\n", *simRequestMsg.WowCharacter.Region)

			performSim(*simRequestMsg.WowCharacter.Region, *simRequestMsg.WowCharacter.Realm, *simRequestMsg.WowCharacter.CharacterName)

			receivedCount += 1
		}
	}()
	<-forever
}
