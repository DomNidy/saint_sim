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
		log.Printf("Failed to execute sim binary: %v", err)
	}

	data, err := os.ReadFile(outputFilePath)
	if err != nil {
		log.Printf("Failed to read sim output file: %v", err)
	}
	fmt.Printf("Sim data: %v\n", data)
}

func main() {
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
			log.Printf("Received a message: %s\n", d.Body)
			log.Printf("receivedCount = %d\n", receivedCount)

			var simRequestMsg interfaces.SimulationMessageBody

			// todo: figure out how to validate the request object
			err := json.Unmarshal(d.Body, &simRequestMsg)
			if err != nil {
				log.Printf("error unmarshalling json: %v", err)
			}
			log.Printf("Received simulation request: %s", string(d.Body))

			// fmt.Printf("	character_name: %s\n", *simRequestMsg.WowCharacter.CharacterName)
			// fmt.Printf("	realm: %s\n", *simRequestMsg.WowCharacter.Realm)
			// fmt.Printf("	region: %s\n", *simRequestMsg.WowCharacter.Region)

			// performSim(*simRequestMsg.WowCharacter.Region, *simRequestMsg.WowCharacter.Realm, *simRequestMsg.WowCharacter.CharacterName)

			receivedCount += 1
		}
	}()
	<-forever
}
