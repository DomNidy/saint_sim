package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	amqp091 "github.com/rabbitmq/amqp091-go"

	"github.com/DomNidy/saint_sim/internal/api"
	"github.com/DomNidy/saint_sim/internal/db"
	"github.com/DomNidy/saint_sim/internal/simc"
	utils "github.com/DomNidy/saint_sim/internal/utils"
)

type simulationWorker struct {
	runner Runner
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
	options, err := request.options.AsSimulationOptionsBasic()
	if err != nil {
		return fmt.Errorf("could not cast to basic: %w", err)
	}

	if err := utils.ValidateSimulationOptionsBasic(&options); err != nil {
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

	run, err := worker.runSimcOnProfile(ctx, options.SimcAddonExport)
	if err != nil {
		return fmt.Errorf("run simulation: %w", err)
	}

	parsed, err := simc.ParseJSON2(run.JSON2)
	if err != nil {
		return fmt.Errorf("parse simc json2 output: %w", err)
	}

	basicResult, err := buildBasicResult(run, parsed)
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
	opts, err := request.options.AsSimulationOptionsTopGear()
	if err != nil {
		return fmt.Errorf("could not cast to topGear: %w", err)
	}

	if opts.Equipment == nil {
		return errTopGearMissingEquipment
	}

	profilesetCount, err := countTopGearProfilesets(opts.Equipment)
	if err != nil {
		return fmt.Errorf("count top gear profilesets: %w", err)
	}

	if profilesetCount > maxGeneratedProfilesets {
		return fmt.Errorf(
			"%w: generated %d, max %d",
			errTopGearProfilesetLimit,
			profilesetCount,
			maxGeneratedProfilesets,
		)
	}

	manifest, err := generateTopGearProfilesets(opts.Equipment, opts.TalentLoadout.Talents)
	if err != nil {
		return fmt.Errorf("generate top gear profilesets: %w", err)
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

	profileText, err := manifest.SimcLines()
	if err != nil {
		return fmt.Errorf("build top gear profile text: %w", err)
	}

	run, err := worker.runSimcOnProfile(ctx, strings.Join(profileText, "\n"))
	if err != nil {
		return fmt.Errorf("run simulation: %w", err)
	}

	parsed, err := simc.ParseJSON2(run.JSON2)
	if err != nil {
		return fmt.Errorf("parse simc json2 output: %w", err)
	}

	topGearResult, err := buildTopGearResult(&manifest, parsed)
	if err != nil {
		return fmt.Errorf("build top gear result: %w", err)
	}

	simResultBytes, err := marshalResultAsTopGear(topGearResult)
	if err != nil {
		return fmt.Errorf("marshal top gear result: %w", err)
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

// runSimcOnProfile is the temp-dir / write / exec dance shared by both
// processBasic and processTopGear. Lives here rather than on Runner because
// it's scoped to how the worker consumes the runner.
func (worker simulationWorker) runSimcOnProfile(
	ctx context.Context,
	profileText string,
) (RunResult, error) {
	tempDir, err := os.MkdirTemp("", "saint-simc-*")
	if err != nil {
		return RunResult{}, fmt.Errorf("create simc temp dir: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			fmt.Fprintf(os.Stderr, "remove simc temp dir: %v\n", removeErr)
		}
	}()

	profilePath := filepath.Join(tempDir, "input.simc")
	if err := os.WriteFile(profilePath, []byte(profileText), simcProfileFileMode); err != nil {
		return RunResult{}, fmt.Errorf("write simc profile: %w", err)
	}

	return worker.runner.Run(ctx, profilePath)
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
