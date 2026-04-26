package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/rabbitmq/amqp091-go"

	"github.com/DomNidy/saint_sim/internal/simulation"
)

const simulationQueueName = "simulation_queue"

type Credentials struct {
	// rabbit mq creds
	User     string
	Password string
	Host     string
	Port     string
}

// SimulationQueue adapts Saint's simulation queue port to RabbitMQ.
type SimulationQueue struct {
	connection *amqp091.Connection
	channel    *amqp091.Channel
	queue      amqp091.Queue
}

// New initializes a connection and channel, ensures the desired rabbitmq messaging
// topology is available (declares queues), and returns a SimulationQueue abstraction.
//
// AMQ declarations are idempotent, meaning they won't create new queues if one already
// exists with the same name & config. If a queue with the same name exists, but has
// a different config, this will error.
func New(creds Credentials) (*SimulationQueue, error) {
	connectionURI := fmt.Sprintf(
		"amqp://%s:%s@%s",
		creds.User,
		creds.Password,
		net.JoinHostPort(creds.Host, creds.Port),
	)

	connection, err := amqp091.Dial(connectionURI)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		_ = connection.Close()

		return nil, fmt.Errorf("open rabbitmq channel: %w", err)
	}

	// declare the simulation queue so we can pub & consume from it
	queue, err := channel.QueueDeclare(
		simulationQueueName,
		false, // durable
		false, // autodelet
		false, // exclusive
		false, // nowait
		nil,
	)
	if err != nil {
		_ = channel.Close()
		_ = connection.Close()

		return nil, fmt.Errorf("declare simulation queue: %w", err)
	}

	return &SimulationQueue{
		connection: connection,
		channel:    channel,
		queue:      queue,
	}, nil
}

func (queue *SimulationQueue) Publish(message simulation.JobMessage) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal simulation job message: %w", err)
	}

	err = queue.channel.Publish(
		"", // exchange name (empty = default exchange)
		queue.queue.Name,
		true,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("publish simulation job message: %w", err)
	}

	return nil
}

func (queue *SimulationQueue) Consume(ctx context.Context) (<-chan amqp091.Delivery, error) {
	deliveries, err := queue.channel.ConsumeWithContext(
		ctx,
		queue.queue.Name,
		"",
		false, // auto ack
		false, // exclusive
		false,
		false,
		amqp091.NewConnectionProperties(),
	)
	if err != nil {
		return nil, fmt.Errorf("consume simulation job messages: %w", err)
	}

	return deliveries, nil
}

func (queue *SimulationQueue) Close() error {
	if queue.channel != nil {
		if err := queue.channel.Close(); err != nil {
			return fmt.Errorf("close rabbitmq channel: %w", err)
		}
	}

	if queue.connection != nil {
		if err := queue.connection.Close(); err != nil {
			return fmt.Errorf("close rabbitmq connection: %w", err)
		}
	}

	return nil
}
