// Package utils contains shared helpers used across services.
package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"

	api_types "github.com/DomNidy/saint_sim/pkg/go-shared/api_types"
	secrets "github.com/DomNidy/saint_sim/pkg/go-shared/secrets"
)

var errUnexpectedUUIDValueType = errors.New("uuid was not encoded as string")

// IntPtr returns a pointer to i.
func IntPtr(i int) *int {
	return &i
}

// StrPtr returns a pointer to s.
func StrPtr(s string) *string {
	return &s
}

// UUIDString converts a Postgres UUID to its string representation.
func UUIDString(id pgtype.UUID) (string, error) {
	value, err := id.Value()
	if err != nil {
		return "", fmt.Errorf("read uuid value: %w", err)
	}

	uuidString, ok := value.(string)
	if !ok {
		return "", errUnexpectedUUIDValueType
	}

	return uuidString, nil
}

// FailOnError logs msg and panics when err is non-nil.
func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s:%s", msg, err)
	}
}

// SimulationQueueClient is a client connection to the simulation queue.
// It can be used by publishers and consumers.
type SimulationQueueClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
	appID   string
}

// NewSimulationQueueClient creates and initializes a simulation queue client.
func NewSimulationQueueClient(
	appID, user, pass, host, port string,
) (*SimulationQueueClient, error) {
	if appID == "" {
		log.Println(
			"WARNING: creating SimulationQueueClient with an empty appID. Avoid this when possible for clarity.",
		)
	}

	var simQueueClient SimulationQueueClient

	simQueueClient.appID = appID

	err := simQueueClient.initialize(user, pass, host, port)
	if err != nil {
		return nil, err
	}

	return &simQueueClient, nil
}

// SimulationMessage is sent to the simulation queue.
type SimulationMessage struct {
	SimulationID string `json:"simulation_id"` //nolint:tagliatelle // external wire format uses snake_case.
}

// Publish writes a simulation message to the queue.
func (s *SimulationQueueClient) Publish(simMsg SimulationMessage) error {
	const (
		mandatory = true
		immediate = false
	)

	jsonSimMsg, err := json.Marshal(simMsg)
	if err != nil {
		return fmt.Errorf("marshal simulation message: %w", err)
	}

	err = s.channel.Publish(
		"",
		s.queue.Name,
		mandatory,
		immediate,
		amqp.Publishing{
			Headers:         nil,
			ContentType:     "application/json",
			ContentEncoding: "",
			DeliveryMode:    0,
			Priority:        0,
			CorrelationId:   "",
			ReplyTo:         "",
			Expiration:      "",
			MessageId:       "",
			Timestamp:       timeZero(),
			Type:            "",
			UserId:          "",
			AppId:           s.appID,
			Body:            jsonSimMsg,
		},
	)
	if err != nil {
		return fmt.Errorf("publish simulation message: %w", err)
	}

	return nil
}

// Consume starts consuming messages from the simulation queue.
func (s *SimulationQueueClient) Consume(
	consumer string,
	autoAck bool,
	exclusive bool,
	noLocal bool,
	noWait bool,
	args amqp.Table,
) (<-chan amqp.Delivery, error) {
	deliveryChannel, err := s.channel.Consume(
		s.queue.Name,
		consumer,
		autoAck,
		exclusive,
		noLocal,
		noWait,
		args,
	)
	if err != nil {
		return nil, fmt.Errorf("consume simulation messages: %w", err)
	}

	return deliveryChannel, nil
}

// ConsumeSimulationMessages consumes using the worker defaults.
func (s *SimulationQueueClient) ConsumeSimulationMessages() (<-chan amqp.Delivery, error) {
	return s.Consume(
		"",
		true,
		false,
		false,
		false,
		nil,
	)
}

// Close closes the queue channel and connection.
func (s *SimulationQueueClient) Close() {
	log.Printf("Closing SimulationQueueClient...")

	if s.channel != nil {
		err := s.channel.Close()
		if err != nil {
			log.Printf("WARNING: failed to close simulation queue channel: %v", err)
		}
	}

	if s.conn != nil {
		err := s.conn.Close()
		if err != nil {
			log.Printf("WARNING: failed to close simulation queue connection: %v", err)
		}
	}
}

// InitPostgresConnectionPool creates a postgres connection pool.
func InitPostgresConnectionPool(ctx context.Context) *pgxpool.Pool {
	dbUser := secrets.LoadSecret("DB_USER").Value()
	dbPassword := secrets.LoadSecret("DB_PASSWORD").Value()
	dbHost := secrets.LoadSecret("DB_HOST").Value()
	dbName := secrets.LoadSecret("DB_NAME").Value()
	dbPort := "5432"

	connectionURI := fmt.Sprintf(
		"postgresql://%s:%s@%s/%s",
		dbUser,
		dbPassword,
		net.JoinHostPort(dbHost, dbPort),
		dbName,
	)

	log.Printf("Connecting to postgres database with name '%s' at %s:%s", dbName, dbHost, dbPort)

	pool, err := pgxpool.New(ctx, connectionURI)
	FailOnError(err, "Failed to create postgres connection")

	return pool
}

// IsValidSimOptions validates the user-provided simulation options.
func IsValidSimOptions(options *api_types.SimulationOptions) bool {
	return isValidInput(options.WowCharacter.CharacterName) &&
		isValidInput(string(options.WowCharacter.Realm)) &&
		isValidInput(string(options.WowCharacter.Region))
}

func isValidInput(input string) bool {
	valid := regexp.MustCompilePOSIX(`^[[:alnum:]_-]+$`)

	return valid.MatchString(input)
}

// IsValidWowRegion reports whether region is in the allowlist.
func IsValidWowRegion(region string) bool {
	switch api_types.WowCharacterRegion(region) {
	case api_types.Us, api_types.Eu, api_types.Tw, api_types.Cn, api_types.Kr:
		return true
	default:
		return false
	}
}

// IsValidWowRealm reports whether realm is in the allowlist.
func IsValidWowRealm(realm string) bool {
	switch api_types.WowCharacterRealm(realm) {
	case api_types.Draenor, api_types.Hydraxis, api_types.Silvermoon, api_types.Thrall:
		return true
	default:
		return false
	}
}

func (s *SimulationQueueClient) initialize(user, pass, host, port string) error {
	connectionURI := fmt.Sprintf("amqp://%s:%s@%s", user, pass, net.JoinHostPort(host, port))

	connection, err := amqp.Dial(connectionURI)
	if err != nil {
		return fmt.Errorf("dial rabbitmq: %w", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}

	queue, err := channel.QueueDeclare(
		"simulation_queue",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare simulation queue: %w", err)
	}

	s.channel = channel
	s.conn = connection
	s.queue = queue

	return nil
}

func timeZero() time.Time {
	return time.Time{}
}
