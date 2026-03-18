#!/bin/bash
# This script generates an API key to authenticate with the `api`
# Inserts the generated API key into the `api_keys` table of the locally running Postgres database

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "$script_dir/.." && pwd)"
env_file="$repo_root/.env"

if [ -f "$env_file" ]; then
    set -a
    . "$env_file"
    set +a
else
    echo "Missing .env at $env_file."
    echo "Create it so this script can read DB_USER and DB_NAME."
    exit 1
fi

if [ -z "$DB_USER" ] || [ -z "$DB_NAME" ]; then
    echo "Error: DB_USER and DB_NAME must be set in $env_file."
    exit 1
fi

# Generate & hash the API key
generated_api_key=$(openssl rand -hex 32)
hashed_api_key=$(echo -n "$generated_api_key" | sha256sum | awk '{print $1}')

# Get the ID of the container running PostgreSQL
postgres_container_id=$(docker ps -f name=postgres -q)

# Check if the PostgreSQL container is running
if [ -z "$postgres_container_id" ]; then
    echo "Error: Could not find Postgres container running locally."
    echo "Hint: Try running the database with ./scripts/local.sh postgres"
    exit 1
fi

# Insert the generated API key into the database
docker exec -it "$postgres_container_id" psql -U "$DB_USER" -d "$DB_NAME" -c \
"INSERT INTO api_keys (api_key, service_name) VALUES ('$hashed_api_key', 'api');"

if [ $? -eq 0 ]; then
    echo "Success: Inserted API key into the database."
    echo "API key: $generated_api_key"
else
    echo "Error: Failed to insert the API key into the database."
    exit 1
fi
