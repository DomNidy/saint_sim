#!/bin/bash

# get the name of the container running postgres
postgres_container_id=$(docker ps -f name=postgres -q)
backup_file_name="schema_backup_$(date +%s)"
echo "Found postgres container id $postgres_container_id"
echo "Creating schema backup script"
docker exec -it $postgres_container_id pg_dump --schema-only --no-owner --no-acl --clean --if-exists -U postgres > $backup_file_name.sql
echo "Created db backup script $backup_file_name.sql"