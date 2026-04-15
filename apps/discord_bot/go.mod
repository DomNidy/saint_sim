module github.com/DomNidy/saint_sim/apps/discord_bot

go 1.21.6

replace (
	github.com/DomNidy/saint_sim/pkg/api_types => ../../pkg/api_types
	github.com/DomNidy/saint_sim/pkg/db => ../../pkg/db
	github.com/DomNidy/saint_sim/pkg/secrets => ../../pkg/secrets
	github.com/DomNidy/saint_sim/pkg/utils => ../../pkg/utils
)

require (
	github.com/DomNidy/saint_sim/pkg/db v0.0.0
	github.com/DomNidy/saint_sim/pkg/secrets v0.0.0
	github.com/DomNidy/saint_sim/pkg/utils v0.0.0
	github.com/bwmarrin/discordgo v0.28.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.1
)

require (
	github.com/DomNidy/saint_sim/pkg/api_types v0.0.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/rabbitmq/amqp091-go v1.10.0 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
)
