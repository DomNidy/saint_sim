// Package main pulls simulation requests from the simulation queue, executes
// them using simc, and writes the results back to the database.
package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	workerusecases "github.com/DomNidy/saint_sim/apps/simulation_worker/usecases"
	simulationpostgres "github.com/DomNidy/saint_sim/internal/simulation/postgres"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config := loadWorkerConfig()

	dependencies, err := setupWorkerDependencies(ctx)
	if err != nil {
		return fmt.Errorf("initialize worker dependencies: %w", err)
	}

	defer dependencies.dbPool.Close()
	defer dependencies.queue.Close()

	simulationRepository := simulationpostgres.NewRepository(dependencies.dbPool)
	processor := workerusecases.NewProcessSimulationUseCase(
		simulationRepository,
		config.simcBinaryPath,
	)

	worker := simulationWorker{
		workerConfig: config,
		processor:    processor,
	}

	err = worker.Start(ctx, dependencies.queue)
	if err != nil {
		return fmt.Errorf("start simulation consumer: %w", err)
	}

	<-ctx.Done()
	log.Printf("simulation worker shutting down")

	return nil
}
