set dotenv-load := true
set shell := ["bash", "-eu", "-o", "pipefail", "-c"]
set script-interpreter := ["bash", "-eu", "-o", "pipefail"]

sqlc_image := "sqlc/sqlc:1.29.0"

[script]
help:
    cat <<"EOF"
    saint_sim command reference

    Getting Started
      just setup
      just db-start
      just db-migrate
      just api-key
      just start

    Setup and lifecycle
      just help                     Show this help
      just setup                    Create .env from .env.example if needed
      just start                    Build and start the full local stack
      just stop                     Stop the full local stack
      just restart <service>        Rebuild and recreate one service
      just logs [service]           Tail logs for a service (default: api)

    Database
      just db-start                 Start or recreate only Postgres
      just db-migrate               Apply all pending Goose migrations
      just db-status                Show Goose migration status
      just db-down [steps]          Roll back migrations (default: 1)
      just db-new <name>            Create a timestamped SQL migration
      just db-schema-backup         Write a schema-only backup file
      just db-reset                 Delete local Postgres and RabbitMQ volumes

    Maintenance
      just api-key                  Generate and insert a local API key
      just codegen [target]         Generate shared code for db and/or api
      just tidy                     Run go mod tidy across all modules
      just doctor                   Check required host tools and setup

    Utility/Debugging:
      just simc                     Run & get a shell in temporary container
                                    with the simc binary preinstalled.

    Services
      api, discord-bot, worker, postgres, pgadmin, rabbitmq (logs only)

    Notes
      codegen targets: db, api
      Host-installed tools required: just, goose, docker, go, npx.
      WSL is recommended on Windows.
      Linux/macOS users can use their normal shell.
    EOF

# Create .env from .env.example if it does not exist.
[script]
setup:
    if [ -f .env ]; then
      echo ".env already exists. Leaving it unchanged."
    else
      cp .env.example .env
      echo "Created .env from .env.example."
    fi
    cat <<'EOF'
    Review these values in .env before continuing:
    - DISCORD_TOKEN
    - APPLICATION_ID
    - SAINT_API_KEY (generated later with `just api-key`)
    EOF

# Build and start the full local stack.
[script]
start:
    echo "Building and starting the local stack..."
    docker compose build
    docker compose up --detach

# Stop the full local stack.
[script]
stop:
    echo "Stopping the local stack..."
    docker compose down

# Rebuild and recreate a single service.
[script]
restart service:
    compose_service=""
    needs_build="false"
    case "{{ service }}" in
      api)
        compose_service="api"
        needs_build="true"
        ;;
      discord-bot)
        compose_service="discord_bot"
        needs_build="true"
        ;;
      worker)
        compose_service="simulation_worker"
        needs_build="true"
        ;;
      postgres)
        compose_service="postgres"
        ;;
      pgadmin)
        compose_service="pgadmin"
        ;;
      *)
        echo "Invalid service: {{ service }}"
        echo "Allowed values: api, discord-bot, worker, postgres, pgadmin"
        exit 1
        ;;
    esac
    docker compose stop "$compose_service" >/dev/null 2>&1 || true
    if [ "$needs_build" = "true" ]; then
      docker compose up --detach --build --force-recreate "$compose_service"
    else
      docker compose up --detach --force-recreate "$compose_service"
    fi

# Tail logs for a service.
[script]
logs service="api":
    compose_service=""
    case "{{ service }}" in
      api)
        compose_service="api"
        ;;
      discord-bot)
        compose_service="discord_bot"
        ;;
      worker)
        compose_service="simulation_worker"
        ;;
      postgres)
        compose_service="postgres"
        ;;
      pgadmin)
        compose_service="pgadmin"
        ;;
      rabbitmq)
        compose_service="rabbitmq"
        ;;
      *)
        echo "Invalid service: {{ service }}"
        echo "Allowed values: api, discord-bot, worker, postgres, pgadmin, rabbitmq"
        exit 1
        ;;
    esac
    docker compose logs --follow --tail=100 "$compose_service"

# Start or recreate only the Postgres service.
[script]
db-start:
    echo "Starting Postgres..."
    docker compose stop postgres >/dev/null 2>&1 || true
    docker compose up --detach --force-recreate postgres

# Apply all pending Goose migrations.
[script]
db-migrate:
    if [ ! -f .env ]; then
      echo "Missing .env. Run \`just setup\` first."
      exit 1
    fi
    : "${DB_USER:?Missing DB_USER in .env}"
    : "${DB_PASSWORD:?Missing DB_PASSWORD in .env}"
    : "${DB_NAME:?Missing DB_NAME in .env}"
    db_port="${DB_PORT:-5432}"
    goose -dir ./db/migrations postgres "user=$DB_USER password=$DB_PASSWORD host=localhost port=$db_port dbname=$DB_NAME sslmode=disable" up

# Show Goose migration status.
[script]
db-status:
    if [ ! -f .env ]; then
      echo "Missing .env. Run \`just setup\` first."
      exit 1
    fi
    : "${DB_USER:?Missing DB_USER in .env}"
    : "${DB_PASSWORD:?Missing DB_PASSWORD in .env}"
    : "${DB_NAME:?Missing DB_NAME in .env}"
    db_port="${DB_PORT:-5432}"
    goose -dir ./db/migrations postgres "user=$DB_USER password=$DB_PASSWORD host=localhost port=$db_port dbname=$DB_NAME sslmode=disable" status

# Roll back Goose migrations by a number of steps.
[script]
db-down steps="1":
    if [ ! -f .env ]; then
      echo "Missing .env. Run \`just setup\` first."
      exit 1
    fi
    : "${DB_USER:?Missing DB_USER in .env}"
    : "${DB_PASSWORD:?Missing DB_PASSWORD in .env}"
    : "${DB_NAME:?Missing DB_NAME in .env}"
    db_port="${DB_PORT:-5432}"
    if ! [[ "{{ steps }}" =~ ^[0-9]+$ ]] || [ "{{ steps }}" -lt 1 ]; then
      echo "steps must be a positive integer."
      exit 1
    fi
    count=0
    while [ "$count" -lt "{{ steps }}" ]; do
      goose -dir ./db/migrations postgres "user=$DB_USER password=$DB_PASSWORD host=localhost port=$db_port dbname=$DB_NAME sslmode=disable" down
      count=$((count + 1))
    done

# Create a new timestamped SQL migration.
[script]
db-new name:
    goose -dir ./db/migrations create "{{ name }}" sql

# Write a schema-only backup file at the repo root.
[script]
db-schema-backup:
    if [ ! -f .env ]; then
      echo "Missing .env. Run \`just setup\` first."
      exit 1
    fi
    : "${DB_USER:?Missing DB_USER in .env}"
    : "${DB_NAME:?Missing DB_NAME in .env}"
    backup_file="schema_backup_$(date +%s).sql"
    echo "Writing schema backup to $backup_file"
    docker compose exec -T postgres pg_dump --schema-only --no-owner --no-acl --clean --if-exists -U "$DB_USER" -d "$DB_NAME" > "$backup_file"

# Delete local Postgres and RabbitMQ volumes after confirmation.
[script]
db-reset:
    echo "Warning: this will delete local Postgres and RabbitMQ data."
    read -r -p "Type 'yes' to continue: " confirm
    if [ "$confirm" != "yes" ]; then
      echo "Cancelled."
      exit 1
    fi
    docker compose down
    docker volume rm -f postgres_data rabbitmq_data >/dev/null 2>&1 || true
    cat <<'EOF'
    Local Postgres and RabbitMQ volumes were removed.
    Recovery steps:
      just db-start
      just db-migrate
      just api-key
      just start
    EOF

# Generate and insert a local API key for discord_bot.
[script]
api-key:
    if [ ! -f .env ]; then
      echo "Missing .env. Run \`just setup\` first."
      exit 1
    fi
    : "${DB_USER:?Missing DB_USER in .env}"
    : "${DB_NAME:?Missing DB_NAME in .env}"
    generated_api_key="$(openssl rand -hex 32)"
    hashed_api_key="$(printf '%s' "$generated_api_key" | openssl dgst -sha256 -r | awk '{print $1}')"
    docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U "$DB_USER" -d "$DB_NAME" -c "INSERT INTO api_keys (api_key, service_name) VALUES ('$hashed_api_key', 'api');"
    echo "Success: inserted API key into the database."
    echo "API key: $generated_api_key"

# Generate shared code for the database and/or OpenAPI schema.
[script]
codegen target="":
    generate_db() {
      # v1.30.0 of sqlc crashes in pgx/os-user lookup when sqlc analyzes database.uri.
      mkdir -p ./pkg/go-shared/db ./pkg/ts-shared/db
      docker pull {{ sqlc_image }}
      docker run --rm \
        -e DB_HOST \
        -e DB_NAME \
        -e DB_USER \
        -e DB_PASSWORD \
        --network saint_network \
        -v "$PWD:/src" \
        {{ sqlc_image }} generate -f /src/db/sqlc.yaml
    }

    generate_api() {
      mkdir -p ./pkg/go-shared/api_types ./pkg/ts-shared/api
      go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.16.3 --generate types,skip-prune -o ./pkg/go-shared/api_types/api_types.gen.go -package api_types ./apps/api/openapi.yaml
      npx --yes openapi-typescript@7.8.0 ./apps/api/openapi.yaml -o ./pkg/ts-shared/api/openapi.gen.ts
    }

    case "{{ target }}" in
      "")
        generate_db
        generate_api
        ;;
      db)
        generate_db
        ;;
      api)
        generate_api
        ;;
      *)
        echo "Invalid codegen target: {{ target }}"
        echo "Allowed values: db, api"
        exit 1
        ;;
    esac

# Run go mod tidy across every Go module in the repository.
[script]
tidy:
    find . -type f -name "go.mod" -print0 | while IFS= read -r -d '' mod_file; do
      module_dir="$(dirname "$mod_file")"
      echo "Tidying $module_dir"
      (
        cd "$module_dir"
        go mod tidy
      )
    done

# Validate required host dependencies and local setup.
[script]
doctor:
    missing="false"
    for cmd in just goose docker go bash npx; do
      if ! command -v "$cmd" >/dev/null 2>&1; then
        echo "Missing required command: $cmd"
        missing="true"
      fi
    done
    if ! docker compose version >/dev/null 2>&1; then
      echo "Docker Compose is not available via \`docker compose\`."
      missing="true"
    fi
    if [ ! -f .env ]; then
      echo "Missing .env. Run \`just setup\` first."
      missing="true"
    fi
    if [ "$missing" = "true" ]; then
      exit 1
    fi
    echo "All required host tools are available."

# Get shell inside of a container with simc preinstalled.
# the version of simc is determined by the version of the
# simc base image that is used. The base image version is

# SIMC_IMAGE environment variable.
[script]
simc:
    if [ ! -f .env ]; then
        echo "Missing .env. Run \`just setup\` first."
        missing="true"
    fi
    simc_image_ver="${SIMC_IMAGE:-}"
    if [ -z "$simc_image_ver" ]; then
      echo ".env file does not have SIMC_IMAGE variable. Defaulting to latest image."
      simc_image_ver="simulationcraftorg/simc:latest"
    fi
    echo "Using simc image: ${simc_image_ver}"
    # Add /app/SimulationCraft to path so we can invoke simc with `simc` (just for convenience)
    docker run --rm -it --entrypoint sh $simc_image_ver -lc 'export PATH="/app/SimulationCraft:$PATH"; exec sh'
