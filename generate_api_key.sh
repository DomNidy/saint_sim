#!/bin/bash
# This script generates an API key to authenticate with the `api`
# Inserts the generated API key into the `api_keys` table of the locally running Postgres database

# Generate & hash the API key
generated_api_key=$(openssl rand -hex 32)
hashed_api_key=$(echo -n "$generated_api_key" | sha256sum | awk '{print $1}')

# Get the ID of the container running PostgreSQL
postgres_container_id=$(docker ps -f name=postgres -q)

# Check if the PostgreSQL container is running
if [ -z "$postgres_container_id" ]; then
    echo "Error: Could not find Postgres container running locally."
    echo "Hint: Try running the database with ./local postgres"
    exit 1
fi

# Insert the generated API key into the database
docker exec -it $postgres_container_id psql -U postgres -c \
"INSERT INTO api_keys (api_key, service_name) VALUES ('$hashed_api_key', 'api');"

if [ $? -eq 0 ]; then
    echo "Success: Inserted API key into the database."
    echo "API key: $generated_api_key"
else
    echo "Error: Failed to insert the API key into the database."
    exit 1
fi