package utils

import (
	"fmt"
	"log"

	secrets "github.com/DomNidy/saint_sim/pkg/secrets"
	// "github.com/jackc/pgx/v5"
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

func InitRabbitMQConnection() (*amqp.Connection, *amqp.Channel) {
	RABBITMQ_USER := secrets.LoadSecret("RABBITMQ_USER")
	RABBITMQ_PASS := secrets.LoadSecret("RABBITMQ_PASS")
	RABBITMQ_PORT := secrets.LoadSecret("RABBITMQ_PORT")
	RABBITMQ_HOST := secrets.LoadSecret("RABBITMQ_HOST")
	connectionURI := fmt.Sprintf("amqp://%s:%s@%s:%s", RABBITMQ_USER.Value(), RABBITMQ_PASS.Value(), RABBITMQ_HOST.Value(), RABBITMQ_PORT.Value())

	conn, err := amqp.Dial(connectionURI)
	FailOnError(err, "Failed to establish RabbitMQ connection")

	// Create channel
	ch, err := conn.Channel()
	FailOnError(err, "Failed to open RabbitMQ channel")

	return conn, ch
}

// func InitPostgresConnection() (*pgx.Conn) {
// 	context.WithTimeout()
// 	DB_USER := secrets.LoadSecret("DB_USER")
// 	DB_PASSWORD := secrets.LoadSecret("DB_PASSWORD")
// 	connectionURI := fmt.Sprint("")
// 	conn,err := pgx.Connect()
// }
