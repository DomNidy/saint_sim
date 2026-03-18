#!/bin/bash

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

# get the name of the container running postgres
postgres_container_id=$(docker ps -f name=postgres -q)
# make sure we were able to find the container
if [ -z "$postgres_container_id" ]; then
    echo "Could not find postgres container running locally, so we cannot create schema backup."
    echo "hint: try running the database with ./scripts/local.sh postgres"
    exit 1
fi

backup_file_name="$repo_root/schema_backup_$(date +%s).sql"
echo "Found postgres container id $postgres_container_id"
echo "Creating schema backup script"
docker exec -i "$postgres_container_id" pg_dump --schema-only --no-owner --no-acl --clean --if-exists -U "$DB_USER" -d "$DB_NAME" > "$backup_file_name"
echo "Created db backup script $backup_file_name"
