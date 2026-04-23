package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	amqp091 "github.com/rabbitmq/amqp091-go"

	"github.com/DomNidy/saint_sim/apps/simulation_worker/json2"
	"github.com/DomNidy/saint_sim/apps/simulation_worker/sims"
	"github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/db"
	utils "github.com/DomNidy/saint_sim/internal/utils"
)

type simulationWorker struct {
	runner sims.Runner
	store  Store
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
	kind, err := request.options.Discriminator()
	if err != nil {
		log.Printf("got error checking for discriminator: %v", err)
		markSimErrored(requestID)
		return
	}

	switch kind {
	case string(api.SimulationKindBasic):
		err = worker.processBasic(ctx, request)
		if err != nil {
			log.Printf("failed to process basic simulation %s: %v", requestID.String(), err)
			markSimErrored(requestID)
		}

		return
	case string(api.SimulationKindTopGear):
		err = worker.processTopGear(ctx, request)
		if err != nil {
			log.Printf("failed to process topGear simulation: %v", err)
			markSimErrored(requestID)
		}

		return
	default:
		log.Printf("got unsupported simulation, ignoring this. kind: '%s'", kind)
		// todo: mark as errored in db
		return
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

// nowTimestamptz returns a pgtype.Timestamptz for the current UTC instant.
func nowTimestamptz() pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:             time.Now().UTC(),
		InfinityModifier: pgtype.Finite,
		Valid:            true,
	}
}

func (worker simulationWorker) processBasic(
	ctx context.Context,
	request simulationRequest,
) error {
	config, err := request.options.AsSimulationConfigBasic()
	if err != nil {
		return fmt.Errorf("could not cast to basic: %w", err)
	}

	if err := utils.ValidateSimulationConfigBasic(&config); err != nil {
		return fmt.Errorf("validate simulation options: %w", err)
	}

	_, err = worker.store.UpdateSimulation(ctx, db.UpdateSimulationParams{
		ID:        request.id,
		StartedAt: nowTimestamptz(),
		Status: db.NullSimulationStatus{
			SimulationStatus: db.SimulationStatusInProgress,
			Valid:            true,
		},
	})
	if err != nil {
		log.Printf("unable to mark simulation %s as started: %v", request.id.String(), err)
	}

	run, err := worker.runner.Run(ctx, config.SimcAddonExport)
	if err != nil {
		return fmt.Errorf("run simulation: %w", err)
	}

	parsed, err := json2.ParseJSON2(run.JSON2)
	if err != nil {
		return fmt.Errorf("parse simc json2 output: %w", err)
	}

	basicResult, err := sims.BuildBasicResult(string(run.Stdout), parsed)
	if err != nil {
		return fmt.Errorf("build basic result: %w", err)
	}

	simResultBytes, err := marshalResultAsBasic(basicResult)
	if err != nil {
		return fmt.Errorf("marshal basic result: %w", err)
	}

	_, err = worker.store.UpdateSimulation(ctx, db.UpdateSimulationParams{
		ID:           request.id,
		SimResult:    simResultBytes,
		SimcRawJson2: run.JSON2,
		CompletedAt:  nowTimestamptz(),
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

func (worker simulationWorker) processTopGear(
	ctx context.Context,
	request simulationRequest,
) error {
	opts, err := request.options.AsSimulationConfigTopGear()
	if err != nil {
		return fmt.Errorf("could not cast to topGear: %w", err)
	}

	manifest, err := sims.NewTopGearManifest(opts)
	if err != nil {
		return err
	}

	_, err = worker.store.UpdateSimulation(ctx, db.UpdateSimulationParams{
		ID:        request.id,
		StartedAt: nowTimestamptz(),
		Status: db.NullSimulationStatus{
			SimulationStatus: db.SimulationStatusInProgress,
			Valid:            true,
		},
	})
	if err != nil {
		log.Printf("unable to mark simulation %s as started: %v", request.id.String(), err)
	}

	simResult, rawJson2, err := manifest.Run(ctx, worker.runner)
	if err != nil {
		return fmt.Errorf("error performing sim gear result: %w", err)
	}

	// if we fail to marshal the raw json 2 into byte array - try to continue still
	var json2Bytes []byte
	json2Bytes, err = json.Marshal(rawJson2)
	if err != nil {
		log.Printf(
			"WARN: failed to marshal rawJson2 into a byte array - db will be missing raw JSON2 output for this sim!",
		)
	}

	// if we fail to marshal the api shape result object, then we need to fail since
	// we have nothing to show the user
	simResultBytes, err := json.Marshal(simResult)
	if err != nil {
		return fmt.Errorf(
			"WARN: failed to marshal the simulation results into a byte array. failing the sim!")
	}

	_, err = worker.store.UpdateSimulation(ctx, db.UpdateSimulationParams{
		ID:           request.id,
		SimResult:    simResultBytes,
		SimcRawJson2: json2Bytes,
		CompletedAt:  nowTimestamptz(),
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

// marshalResultAsBasic wraps a basic result in the SimulationResult union
// and returns the canonical on‑wire / on‑disk bytes. The union's custom
// MarshalJSON emits only the inner variant (with kind embedded) — that's
// exactly the shape we want in the `sim_result` column.
func marshalResultAsBasic(r api.SimulationResultBasic) ([]byte, error) {
	var u api.SimulationResult
	if err := u.FromSimulationResultBasic(r); err != nil {
		return nil, fmt.Errorf("wrap basic result: %w", err)
	}
	return json.Marshal(u)
}

func marshalResultAsTopGear(r api.SimulationResultTopGear) ([]byte, error) {
	var u api.SimulationResult
	if err := u.FromSimulationResultTopGear(r); err != nil {
		return nil, fmt.Errorf("wrap top gear result: %w", err)
	}
	return json.Marshal(u)
}
