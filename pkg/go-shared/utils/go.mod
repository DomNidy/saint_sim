module github.com/DomNidy/saint_sim/pkg/go-shared/utils

go 1.21.6

replace (
	github.com/DomNidy/saint_sim/pkg/go-shared/api_types => ../api_types
	github.com/DomNidy/saint_sim/pkg/go-shared/secrets => ../secrets
)

require (
	github.com/DomNidy/saint_sim/pkg/go-shared/api_types v0.0.0-20260404192636-ff4dbb6469b9
	github.com/DomNidy/saint_sim/pkg/go-shared/secrets v0.0.0-20260404192636-ff4dbb6469b9
	github.com/jackc/pgx/v5 v5.7.1
	github.com/rabbitmq/amqp091-go v1.10.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/text v0.19.0 // indirect
)
