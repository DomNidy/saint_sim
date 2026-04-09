package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	amqp091 "github.com/rabbitmq/amqp091-go"

	utils "github.com/DomNidy/saint_sim/pkg/go-shared/utils"
)

type simulationWorker struct {
	runner simcRunner
	store  simulationStore
}

func (worker simulationWorker) Start(
	ctx context.Context,
	queue *utils.SimulationQueueClient,
) error {
	msgChan, err := queue.ConsumeSimulationMessages()
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

	requestIDText, requestID, err := parseSimulationRequestID(msg)
	if err != nil {
		log.Printf("discarding malformed simulation message: %v", err)

		return
	}

	request, err := worker.store.LoadRequest(ctx, requestID, requestIDText)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("simulation %s no longer exists; skipping message", requestIDText)
		} else {
			log.Printf("failed to load simulation %s: %v", requestIDText, err)
		}

		return
	}

	err = worker.processRequest(ctx, request)
	if err != nil {
		log.Printf("simulation %s failed: %v", request.idText, err)

		markErr := worker.store.MarkFailed(ctx, request.id, err)
		if markErr != nil {
			log.Printf(
				"failed to persist simulation error for %s: %v",
				request.idText,
				markErr,
			)
		}
	}
}

func (worker simulationWorker) processRequest(
	ctx context.Context,
	request simulationRequest,
) error {
	target, err := simulationTargetFromOptions(request.options)
	if err != nil {
		return err
	}

	err = worker.store.MarkStarted(ctx, request.id)
	if err != nil {
		log.Printf("unable to mark simulation %s as started: %v", request.idText, err)
	}

	result, err := worker.runner.Run(ctx, target)
	if err != nil {
		return fmt.Errorf("run simulation: %w", err)
	}

	err = worker.store.MarkCompleted(ctx, request.id, result)
	if err != nil {
		return fmt.Errorf("persist simulation result: %w", err)
	}

	return nil
}

func parseSimulationRequestID(msg amqp091.Delivery) (string, pgtype.UUID, error) {
	var simRequestMsg utils.SimulationMessage

	err := json.Unmarshal(msg.Body, &simRequestMsg)
	if err != nil {
		return "", pgtype.UUID{}, fmt.Errorf("unmarshal simulation message: %w", err)
	}

	var requestID pgtype.UUID

	err = requestID.Scan(simRequestMsg.SimulationID)
	if err != nil {
		return "", pgtype.UUID{}, fmt.Errorf(
			"parse simulation id %q: %w",
			simRequestMsg.SimulationID,
			err,
		)
	}

	return simRequestMsg.SimulationID, requestID, nil
}
