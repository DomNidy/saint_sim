package utils

// package utils

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"time"

// 	amqp "github.com/rabbitmq/amqp091-go"
// )

// // SimulationQueueClient is a client connection to the simulation queue.
// // It can be used by publishers and consumers.
// type SimulationQueueClient struct {
// 	conn    *amqp.Connection
// 	channel *amqp.Channel
// 	queue   amqp.Queue
// 	appID   string
// }

// // SimulationJobMessage is sent to the simulation queue.
// type SimulationJobMessage struct {
// 	SimulationID string `json:"simulation_id"` //nolint:tagliatelle // external wire format uses snake_case.
// }

// // NewSimulationQueueClient creates and initializes a simulation queue client.
// func NewSimulationQueueClient(
// 	appID, user, pass, host, port string,
// ) (*SimulationQueueClient, error) {
// 	if appID == "" {
// 		log.Println(
// 			"WARNING: creating SimulationQueueClient with an empty appID. Avoid this when possible for
// clarity.",
// 		)
// 	}

// 	var simQueueClient SimulationQueueClient

// 	simQueueClient.appID = appID

// 	err := simQueueClient.initialize(user, pass, host, port)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &simQueueClient, nil
// }

// // Publish writes a simulation message to the queue.
// func (s *SimulationQueueClient) Publish(simMsg SimulationJobMessage) error {
// 	const (
// 		mandatory = true
// 		immediate = false
// 	)

// 	jsonSimMsg, err := json.Marshal(simMsg)
// 	if err != nil {
// 		return fmt.Errorf("marshal simulation message: %w", err)
// 	}

// 	err = s.channel.Publish(
// 		"",
// 		s.queue.Name,
// 		mandatory,
// 		immediate,
// 		amqp.Publishing{
// 			Headers:         nil,
// 			ContentType:     "application/json",
// 			ContentEncoding: "",
// 			DeliveryMode:    0,
// 			Priority:        0,
// 			CorrelationId:   "",
// 			ReplyTo:         "",
// 			Expiration:      "",
// 			MessageId:       "",
// 			Timestamp:       timeZero(),
// 			Type:            "",
// 			UserId:          "",
// 			AppId:           s.appID,
// 			Body:            jsonSimMsg,
// 		},
// 	)
// 	if err != nil {
// 		return fmt.Errorf("publish simulation message: %w", err)
// 	}

// 	return nil
// }

// // Consume starts consuming messages from the simulation queue.
// func (s *SimulationQueueClient) Consume(
// 	consumer string,
// 	autoAck bool,
// 	exclusive bool,
// 	noLocal bool,
// 	noWait bool,
// 	args amqp.Table,
// ) (<-chan amqp.Delivery, error) {
// 	deliveryChannel, err := s.channel.Consume(
// 		s.queue.Name,
// 		consumer,
// 		autoAck,
// 		exclusive,
// 		noLocal,
// 		noWait,
// 		args,
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("consume simulation messages: %w", err)
// 	}

// 	return deliveryChannel, nil
// }

// // ConsumeSimulationJobMessages consumes using the worker defaults.
// func (s *SimulationQueueClient) ConsumeSimulationJobMessages() (<-chan amqp.Delivery, error) {
// 	return s.Consume(
// 		"",
// 		true,
// 		false,
// 		false,
// 		false,
// 		nil,
// 	)
// }

// func timeZero() time.Time {
// 	return time.Time{}
// }
