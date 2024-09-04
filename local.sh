#!/bin/bash

mode=$1

if [[ "$mode" == "start" ]]; then
    echo "Building Docker images..."
    docker-compose build

    echo "Starting saint_sim containers..."
    docker-compose up

    echo "Local environment is ready!"
elif [[ "$mode" == "stop" ]]; then
    echo "Stopping saint_sim containers..."
    docker-compose down
else
    echo "Invalid mode specified. Please specify 'start' or 'stop' in the first argument"
fi