#!/bin/bash
# Generates an API key that can be used to authenticate with the `api`
# automatically inserts this into the `api_keys` table of the locally running postgres database
# get the name of the container running postgres
# requires apache2-utils package (since we use bcrypt to hash pw)
generated_api_key=$(openssl rand -hex 32)
hashed_api_key=$(echo -n "$generated_api_key" | sha256sum | awk '{print $1}')
postgres_container_id=$(docker ps -f name=postgres -q)
if [ -z "$postgres_container_id" ]; then
    echo "Could not find postgres container running locally, so we cannot insert the generated API key."
    echo "hint: try running the database with ./local postgres"
    exit 1
fi
docker exec -it $postgres_container_id psql -U postgres -c "INSERT INTO api_keys (api_key, service_name) VALUES ('$hashed_api_key', 'api');" 
echo "Inserted API key into the database"
echo "API key: $generated_api_key"
