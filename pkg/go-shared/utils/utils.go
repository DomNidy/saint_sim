package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	api_types "github.com/DomNidy/saint_sim/pkg/go-shared/api_types"
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

// SimulationQueueClient a client connection to the simulation queue
// Client can be a consumer or a publisher of messages:
//
//	Backend workers -> consuming client
//	API -> publishing client
type SimulationQueueClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue

	appId string
}

// Create and initialize a new client connection to the
// simulation queue. The appId will be written to the
// body of messages published by this client.
func NewSimulationQueueClient(appId, user, pass, host, port string) (*SimulationQueueClient, error) {
	if appId == "" {
		log.Println("WARNING: creating SimulationQueueClient with an empty appId. Probably should avoid this for clarity.")
	}
	simQueueClient := SimulationQueueClient{appId: appId}
	err := simQueueClient.initialize(user, pass, host, port)
	if err != nil {
		return nil, err
	}
	return &simQueueClient, nil
}

// Connect to server, open a channel, and declare queue
func (s *SimulationQueueClient) initialize(user, pass, host, port string) error {
	connectionURI := fmt.Sprintf("amqp://%s:%s@%s:%s", user, pass, host, port)
	conn, err := amqp.Dial(connectionURI)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	q, err := ch.QueueDeclare(
		"simulation_queue", // name
		false,              // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return err
	}
	s.channel = ch
	s.conn = conn
	s.queue = q
	return nil
}

// Message sent to the simulation queue
type SimulationMessage struct {
	SimulationID string `json:"simulation_id"`
}

// Publish a sim message into the simulation queue
func (s *SimulationQueueClient) Publish(simMsg SimulationMessage) error {
	const mandatory = true  // queue must be bound that matches routing key
	const immediate = false // do not need to **immediately** deliver this to a consumer on the queue

	jsonSimMsg, err := json.Marshal(simMsg)
	if err != nil {
		return err
	}

	err = s.channel.Publish(
		"", // default exchange name
		s.queue.Name,
		mandatory,
		immediate,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonSimMsg,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// Consume immedaitely started delivered queued messages on the channel
func (s *SimulationQueueClient) Consume(
	consumer string,
	autoAck bool,
	exclusive bool,
	noLocal bool,
	noWait bool,
	args amqp.Table,
) (<-chan amqp.Delivery, error) {
	channel, err := s.channel.Consume(
		s.queue.Name,
		consumer,
		autoAck,
		exclusive,
		noLocal,
		noWait,
		args,
	)

	return channel, fmt.Errorf("%w: SimulationQueueClient error consuming from queue", err)
}

func (s *SimulationQueueClient) Close() {
	log.Printf("Closing SimulationQueueClient...")
	if s.channel != nil {
		s.channel.Close()
	}
	if s.conn != nil {
		s.conn.Close()
	}
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

func IsValidSimOptions(options *api_types.SimulationOptions) bool {
	if !isValidInput(options.WowCharacter.CharacterName) ||
		!isValidInput(string(options.WowCharacter.Realm)) ||
		!isValidInput(string(options.WowCharacter.Region)) {
		return false
	}

	return true
}

// * Important
// Utility to validate command line arguments received before they are used to execute the simc command on the sim worker
// Also use this in other places where the user input is passed, like at the api, discord bot, etc.
// allows alphanumeric chars, and underscores (underscores are safe to allow, right?)
func isValidInput(input string) bool {
	valid := regexp.MustCompilePOSIX(`^[[:alnum:]_-]+$`)
	return valid.MatchString(input)
}

func IsValidWowRegion(region string) bool {
	_, exists := validWowRegions[api_types.WowCharacterRegion(region)]
	return exists
}

func IsValidWowRealm(realm string) bool {
	_, exists := validWowRealms[api_types.WowCharacterRealm(realm)]
	return exists
}

var validWowRealms = map[api_types.WowCharacterRealm]struct{}{
	api_types.Draenor:    {},
	api_types.Hydraxis:   {},
	api_types.Silvermoon: {},
	api_types.Thrall:     {},
}

var validWowRegions = map[api_types.WowCharacterRegion]struct{}{
	api_types.Us: {},
	api_types.Eu: {},
	api_types.Tw: {},
	api_types.Cn: {},
	api_types.Kr: {},
}
