package utils

import (
	"context"
	"fmt"
	"log"

	secrets "github.com/DomNidy/saint_sim/pkg/go-shared/secrets"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

// todo: Does go gc will clean this up, right?: https://tip.golang.org/doc/gc-guide
// todo: also, is this safe? aren't we assigning i to the memory that gets allocated
// todo: for the function params, which, is local to the stack frame of this func?
// todo: idk, seems like the returned values are valid anyway.
// Helper function used to concisely 'inline' an int pointer
func IntPtr(i int) *int {
	return &i
}

func StrPtr(s string) *string {
	return &s
}

// Utility function used to open up rabbitmq connections, channels, queues, etc. easier
func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s:%s", msg, err)
	}
}

func SanitizeAndParseSimcString(simcConfigString string) (string, error) {
	return "", nil
}

// Creates a rabbit mq channel with a single connection
// A channel multiplexes connections over a single TCP connection
// This allows us to logically distinguish between different 'connections',
// while ony needing a single TCP connection
func InitRabbitMQConnection() (*amqp.Connection, *amqp.Channel) {
	RABBITMQ_USER := secrets.LoadSecret("RABBITMQ_USER")
	RABBITMQ_PASS := secrets.LoadSecret("RABBITMQ_PASS")
	RABBITMQ_PORT := secrets.LoadSecret("RABBITMQ_PORT")
	RABBITMQ_HOST := secrets.LoadSecret("RABBITMQ_HOST")
	connectionURI := fmt.Sprintf("amqp://%s:%s@%s:%s", RABBITMQ_USER.Value(), RABBITMQ_PASS.Value(), RABBITMQ_HOST.Value(), RABBITMQ_PORT.Value())

	log.Printf("Connecting to RabbitMQ: user=%s, password=%s, host=%s, port=%s", RABBITMQ_USER.MaskedValue(), RABBITMQ_PASS.MaskedValue(), RABBITMQ_HOST.MaskedValue(), RABBITMQ_PORT.MaskedValue())

	conn, err := amqp.Dial(connectionURI)
	FailOnError(err, "Failed to establish RabbitMQ connection")

	// Create channel
	ch, err := conn.Channel()
	FailOnError(err, "Failed to open RabbitMQ channel")

	return conn, ch
}

// Declare simulation queue in channel (creates it if it doesn't exist)
func DeclareSimulationQueue(ch *amqp.Channel) *amqp.Queue {
	q, err := ch.QueueDeclare(
		"simulation_queue", // name
		false,              // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	FailOnError(err, "Failed to declare simulation_queue for channel")
	return &q
}

// Create a postgres connection pool
// This is concurrency safe
func InitPostgresConnectionPool(ctx context.Context) *pgxpool.Pool {
	DB_USER := secrets.LoadSecret("DB_USER").Value()
	DB_PASSWORD := secrets.LoadSecret("DB_PASSWORD").Value()
	DB_HOST := secrets.LoadSecret("DB_HOST").Value()
	DB_NAME := secrets.LoadSecret("DB_NAME").Value()
	DB_PORT := "5432"
	connectionURI := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME)
	log.Printf("Connecting to postgres database with name '%s' at %s:%s", DB_NAME, DB_HOST, DB_PORT)

	pool, err := pgxpool.New(ctx, connectionURI)
	FailOnError(err, "Failed to create postgres connection")
	return pool
}
