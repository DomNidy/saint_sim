#!/bin/bash

mode=$1

if [[ "$mode" == "start" ]]; then
    echo "Building Docker images..."
    docker compose build

    echo "Starting saint_sim containers..."
    echo "$(date)"
    docker compose up --detach
    echo "$(date)"
elif [[ "$mode" == "stop" ]]; then
    echo "Stopping saint_sim containers..."
    docker compose down
elif [[ "$mode" == "api" ]]; then
    echo "Rebuilding and starting api"
    docker compose stop api
    docker compose up --detach --build --force-recreate api
elif [[ "$mode" == "discord_bot" ]]; then
    echo "Rebuilding and starting discord_bot"
    docker compose stop discord_bot
    docker compose up --detach --build --force-recreate discord_bot
elif [[ "$mode" == "simulation_worker" ]]; then
    echo "Rebuilding and simulation_worker"
    docker compose stop simulation_worker
    docker compose up --detach --build --force-recreate simulation_worker
else
    echo "Invalid mode or service specified. Please specify 'start' or 'stop' or '[service_name]' in the first argument"
fi
