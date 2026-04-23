package rabbitmq

import (
	"fmt"
	"net"

	"github.com/rabbitmq/amqp091-go"
)

type Credentials struct {
	// rabbit mq creds
	User     string
	Password string
	Host     string
	Port     string
}

func New(creds Credentials) (*amqp091.Channel, error) {
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
		return nil, fmt.Errorf("open rabbitmq channel: %w", err)
	}

	// declare the simulation queue so we can pub & consume from it
	_, err = channel.QueueDeclare(
		"simulation_queue",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("declare simulation queue: %w", err)
	}

	return channel, nil
}
