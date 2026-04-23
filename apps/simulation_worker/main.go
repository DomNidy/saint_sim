// Package main pulls simulation requests from the simulation queue, executes
// them using simc, and writes the results back to the database.
package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rabbitmq/amqp091-go"

	workerusecases "github.com/DomNidy/saint_sim/apps/simulation_worker/usecases"
	"github.com/DomNidy/saint_sim/internal/platform/postgres"
	"github.com/DomNidy/saint_sim/internal/platform/rabbitmq"
	secrets "github.com/DomNidy/saint_sim/internal/secrets"
	simulationpostgres "github.com/DomNidy/saint_sim/internal/simulation/postgres"
)

type workerConfig struct {
	simcBinaryPath string
	dbPool         *pgxpool.Pool
	rabbitChannel  *amqp091.Channel
}

func loadWorkerConfig(ctx context.Context) (workerConfig, error) {
	pool, err := postgres.New(context.Background(), postgres.Credentials{
		DBUser:     secrets.LoadSecret("DB_USER").Value(),
		DBPassword: secrets.LoadSecret("DB_PASSWORD").Value(),
		DBHost:     secrets.LoadSecret("DB_HOST").Value(),
		DBName:     secrets.LoadSecret("DB_NAME").Value(),
		DBPort:     "5432",
	})

	if err != nil {
		log.Panicf("%s: could not make postgres pool", err)
	}

	rabbitChannel, err := rabbitmq.New(
		rabbitmq.Credentials{
			User:     secrets.LoadSecret("RABBITMQ_USER").Value(),
			Password: secrets.LoadSecret("RABBITMQ_PASS").Value(),
			Host:     secrets.LoadSecret("RABBITMQ_HOST").Value(),
			Port:     secrets.LoadSecret("RABBITMQ_PORT").Value(),
		},
	)

	if err != nil {
		log.Panicf("ERROR: Failed to initialize connection to simulation queue: %v", err)

		return workerConfig{}, nil
	}

	return workerConfig{
		simcBinaryPath: secrets.LoadSecret("SIMC_BINARY_PATH").Value(),
		dbPool:         pool,
		rabbitChannel:  rabbitChannel,
	}, nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config, err := loadWorkerConfig(ctx)
	if err != nil {
		return err
	}

	defer config.dbPool.Close()
	defer config.rabbitChannel.Close()

	simulationRepository := simulationpostgres.NewRepository(config.dbPool)
	processor := workerusecases.NewProcessSimulationUseCase(
		simulationRepository,
		config.simcBinaryPath,
	)

	worker := simulationWorker{
		workerConfig: config,
		processor:    processor,
	}

	simChan, err := config.rabbitChannel.ConsumeWithContext(ctx,
		"simulation_queue",
		"",
		false, // autoack
		false, // exclusive
		false,
		false,
		amqp091.NewConnectionProperties(),
	)

	if err != nil {
		return fmt.Errorf("fail to start consuming from queue : %w", err)
	}

	go worker.consumeLoop(ctx, simChan)

	<-ctx.Done()
	log.Printf("simulation worker shutting down")

	return nil
}
