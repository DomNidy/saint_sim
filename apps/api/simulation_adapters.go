package api

import (
	"context"
	"encoding/json"
	"time"

	api_utils "github.com/DomNidy/saint_sim/apps/api/api_utils"
	api_types "github.com/DomNidy/saint_sim/pkg/go-shared/api_types"
	amqp "github.com/rabbitmq/amqp091-go"
)

type liveCharacterLookup struct{}

func (liveCharacterLookup) Exists(character *api_types.WowCharacter) (bool, error) {
	return api_utils.CheckWowCharacterExists(character)
}

type rabbitMQDispatcher struct {
	channel   *amqp.Channel
	queueName string
	timeout   time.Duration
}

func (d rabbitMQDispatcher) DispatchSimulation(ctx context.Context, msg api_types.SimulationMessageBody) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	publishCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	return d.channel.PublishWithContext(
		publishCtx,
		"", // exchange
		d.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
