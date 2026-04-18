package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	pgx "github.com/jackc/pgx/v5"
	amqp091 "github.com/rabbitmq/amqp091-go"

	"github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/db"
	utils "github.com/DomNidy/saint_sim/internal/utils"
)

type simulationWorker struct {
	runner simcRunner
	store  SimulationStore
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

	request, err := worker.store.LoadRequest(ctx, requestID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("simulation %s no longer exists; skipping message", requestID.String())
		} else {
			log.Printf("failed to load simulation %s: %v", requestID.String(), err)
		}

		return
	}

	markSimErrored := func(simulationId uuid.UUID) {
		worker.store.UpdateSimulation(ctx, db.UpdateSimulationParams{
			ID:        simulationId,
			ErrorText: utils.StrPtr("internal server error"),
			Status: db.NullSimulationStatus{
				SimulationStatus: db.SimulationStatusError, Valid: true,
			},
		})
	}

	// route request to the handler for the simulation kind
	switch request.options.Kind {
	case api.SimulationOptionsKindBasic:
		err = worker.processBasic(ctx, request)
		if err != nil {
			log.Printf("failed to process basic simulation %s: %v", requestID.String(), err)
			markSimErrored(requestID)
		}

		return
	case api.SimulationOptionsKindTopGear:
		err = worker.processTopGear(ctx, request)
		if err != nil {
			log.Printf("got error cast to basic sim options, ignoring this job: %v", err)
			markSimErrored(requestID)
		}

		return
	default:
		log.Printf("got unsupported simulation, ignoring this. kind: '%s'", request.options.Kind)
		// todo: mark as errored in db
		return
	}
}

func (worker simulationWorker) processBasic(
	ctx context.Context,
	request simulationRequest,
) error {
	options, err := request.options.AsSimulationOptionsBasic()
	if err != nil {
		return fmt.Errorf("could not cast to basic: %w", err)
	}

	err = utils.ValidateSimulationOptionsBasic(&options)
	if err != nil {
		return fmt.Errorf("validate simulation options: %w", err)
	}

	_, err = worker.store.UpdateSimulation(ctx, db.UpdateSimulationParams{
		ID: request.id,
		Status: db.NullSimulationStatus{
			SimulationStatus: db.SimulationStatusInProgress,
			Valid:            true,
		},
	})
	if err != nil {
		log.Printf("unable to mark simulation %s as started: %v", request.id.String(), err)
	}

	tempDir, err := os.MkdirTemp("", "saint-simc-*")
	if err != nil {
		return fmt.Errorf("create simc temp dir: %w", err)
	}

	defer func() {
		removeErr := os.RemoveAll(tempDir)
		if removeErr != nil {
			fmt.Fprintf(os.Stderr, "remove simc temp dir: %v\n", removeErr)
		}
	}()

	profilePath := filepath.Join(tempDir, "input.simc")

	err = os.WriteFile(profilePath, []byte(options.SimcAddonExport), simcProfileFileMode)
	if err != nil {
		return fmt.Errorf("write simc profile: %w", err)
	}

	result, err := worker.runner.Run(ctx, profilePath)
	if err != nil {
		return fmt.Errorf("run simulation: %w", err)
	}

	simResult := string(result)

	_, err = worker.store.UpdateSimulation(ctx, db.UpdateSimulationParams{
		ID:          request.id,
		SimResult:   &simResult,
		CompletedAt: timestampValue(time.Now().UTC()),
		Status: db.NullSimulationStatus{
			SimulationStatus: db.SimulationStatusComplete,
			Valid:            true,
		},
	})
	if err != nil {
		return fmt.Errorf("persist simulation result: %w", err)
	}

	return nil
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

func (worker simulationWorker) processTopGear(
	ctx context.Context,
	request simulationRequest,
) error {
	_, err := request.options.AsSimulationOptionsTopGear()
	if err != nil {
		return fmt.Errorf("could not cast to topGear: %w", err)
	}
	return nil
}
