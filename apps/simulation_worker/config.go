package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	secrets "github.com/DomNidy/saint_sim/pkg/go-shared/secrets"
	utils "github.com/DomNidy/saint_sim/pkg/go-shared/utils"
)

type workerConfig struct {
	simcBinaryPath string
}

type workerDependencies struct {
	dbPool *pgxpool.Pool
	queue  *utils.SimulationQueueClient
}

func loadWorkerConfig() workerConfig {
	return workerConfig{
		simcBinaryPath: secrets.LoadSecret("SIMC_BINARY_PATH").Value(),
	}
}

func setupSimulationQueueConnection() (*utils.SimulationQueueClient, error) {
	user := secrets.LoadSecret("RABBITMQ_USER").Value()
	pass := secrets.LoadSecret("RABBITMQ_PASS").Value()
	host := secrets.LoadSecret("RABBITMQ_HOST").Value()
	port := secrets.LoadSecret("RABBITMQ_PORT").Value()

	queue, err := utils.NewSimulationQueueClient("saint_api", user, pass, host, port)
	if err != nil {
		return nil, fmt.Errorf("initialize simulation queue connection: %w", err)
	}

	return queue, nil
}

func setupWorkerDependencies(ctx context.Context) (workerDependencies, error) {
	queue, err := setupSimulationQueueConnection()
	if err != nil {
		return workerDependencies{}, err
	}

	dbPool, err := utils.InitPostgresConnectionPool(ctx)
	if err != nil {
		queue.Close()

		return workerDependencies{}, fmt.Errorf("initialize postgres connection pool: %w", err)
	}

	return workerDependencies{
		dbPool: dbPool,
		queue:  queue,
	}, nil
}
