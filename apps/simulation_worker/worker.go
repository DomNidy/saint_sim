package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	amqp091 "github.com/rabbitmq/amqp091-go"

	workerusecases "github.com/DomNidy/saint_sim/apps/simulation_worker/usecases"
	"github.com/DomNidy/saint_sim/internal/simulation"
	utils "github.com/DomNidy/saint_sim/internal/utils"
)

type simulationProcessor interface {
	Process(ctx context.Context, requestID uuid.UUID) error
}

type simulationWorker struct {
	workerConfig
	processor simulationProcessor
}

func (worker simulationWorker) Start(
	ctx context.Context,
	queue *utils.SimulationQueueClient,
) error {
	msgChan, err := queue.ConsumeSimulationJobMessages()
	if err != nil {
		return fmt.Errorf("register as simulation consumer: %w", err)
	}

	go worker.consumeLoop(ctx, msgChan)

	return nil
}

func (worker simulationWorker) consumeLoop(ctx context.Context, msgChan <-chan amqp091.Delivery) {
	var receivedCount uint64

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgChan:
			if !ok {
				log.Printf("simulation consumer channel closed")

				return
			}

			receivedCount++
			worker.handleDelivery(ctx, msg, receivedCount)
		}
	}
}

func (worker simulationWorker) handleDelivery(
	ctx context.Context,
	msg amqp091.Delivery,
	receivedCount uint64,
) {
	log.Printf("received simulation message #%d: %s", receivedCount, string(msg.Body))

	requestID, err := parseSimulationRequestID(msg)
	if err != nil {
		log.Printf("discarding malformed simulation message: %v", err)

		return
	}

	err = worker.processor.Process(ctx, requestID)
	if err == nil {
		return
	}

	switch {
	case errors.Is(err, simulation.ErrNotFound):
		log.Printf("simulation %s no longer exists; skipping message", requestID.String())
	case errors.Is(err, workerusecases.ErrUnsupportedSimulationKind):
		log.Printf("got unsupported simulation, ignoring this: %v", err)
	default:
		log.Printf("failed to process simulation %s: %v", requestID.String(), err)
	}
}

func parseSimulationRequestID(msg amqp091.Delivery) (uuid.UUID, error) {
	var simRequestMsg utils.SimulationJobMessage

	err := json.Unmarshal(msg.Body, &simRequestMsg)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("unmarshal simulation message: %w", err)
	}

	requestID, err := uuid.Parse(simRequestMsg.SimulationID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf(
			"parse simulation id %q: %w",
			simRequestMsg.SimulationID,
			err,
		)
	}

	return requestID, nil
}
