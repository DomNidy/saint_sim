// Package main runs the Gin API server.
package main

import (
	"context"
	"log"
	"time"

	"github.com/go-jose/go-jose/v4/jwt"

	"github.com/DomNidy/saint_sim/apps/api/auth"
	handlers "github.com/DomNidy/saint_sim/apps/api/handlers"
	api "github.com/DomNidy/saint_sim/internal/api"
	dbqueries "github.com/DomNidy/saint_sim/internal/db"
	"github.com/DomNidy/saint_sim/internal/platform/postgres"
	"github.com/DomNidy/saint_sim/internal/platform/rabbitmq"
	"github.com/DomNidy/saint_sim/internal/secrets"
	simulationpostgres "github.com/DomNidy/saint_sim/internal/simulation/postgres"
)

func newJWTAuthenticator(
	dbClient *dbqueries.Queries,
	betterAuthURL string,
) *auth.JWTAuthenticator {
	return auth.NewJWTAuthenticator(dbClient, &jwt.Expected{
		// only accept jwt issued by our web app
		Issuer: betterAuthURL,
		// dont accept allow tokens for other audiences
		AnyAudience: []string{"saint-api"},
		Subject:     "",
		ID:          "",
		Time:        time.Time{},
	})
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile) // include filename:linenumber in log output

	betterAuthURL := secrets.LoadSecret("BETTER_AUTH_URL").Value()
	user := secrets.LoadSecret("RABBITMQ_USER").Value()
	pass := secrets.LoadSecret("RABBITMQ_PASS").Value()
	host := secrets.LoadSecret("RABBITMQ_HOST").Value()
	port := secrets.LoadSecret("RABBITMQ_PORT").Value()

	simulationQueue, err := rabbitmq.New(
		rabbitmq.Credentials{User: user, Password: pass, Host: host, Port: port},
	)
	if err != nil {
		log.Panicf("ERROR: Failed to initialize connection to simulation queue: %v", err)

		return
	}
	defer simulationQueue.Close()

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
	defer pool.Close()

	dbClient := dbqueries.New(pool)
	if dbClient == nil {
		log.Panicf("API failed to acquire a dbClient, cannot run.")
	}

	swagger, err := api.GetSwagger()
	if err != nil {
		log.Panicf("ERROR: Failed to load embedded OpenAPI spec: %v", err)
	}

	apiKeyAuthenticator := auth.NewAPIKeyAuthenticator(dbClient)
	jwtAuthenticator := newJWTAuthenticator(dbClient, betterAuthURL)
	openAPIAuthenticator := auth.NewOpenAPIRequestAuthenticator(
		jwtAuthenticator,
		apiKeyAuthenticator,
	)

	simulationRepository := simulationpostgres.NewRepository(pool)
	router := newRouter(
		handlers.NewServer(simulationRepository, simulationQueue),
		swagger,
		openAPIAuthenticator,
	)

	err = router.Run("0.0.0.0:8080")
	if err != nil {
		log.Printf("ERROR: Failed to start API server: %v", err)
	}
}
