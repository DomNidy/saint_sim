// Package main pulls simulation requests from the simulation queue, executes
// them using simc, and writes the results back to the database.
package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	dbqueries "github.com/DomNidy/saint_sim/pkg/db"
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

	runner := simcRunner{binaryPath: config.simcBinaryPath}

	version, err := runner.Version(ctx)
	if err != nil {
		return fmt.Errorf("read SimC version: %w", err)
	}

	log.Printf("simulation worker running SimC version: %s", version)

	worker := simulationWorker{
		runner: simcRunner{binaryPath: config.simcBinaryPath},
		store: simulationStore{
			queries: *dbqueries.New(dependencies.dbPool),
		},
	}

	err = worker.Start(ctx, dependencies.queue)
	if err != nil {
		return fmt.Errorf("start simulation consumer: %w", err)
	}

	<-ctx.Done()
	log.Printf("simulation worker shutting down")

	return nil
}
