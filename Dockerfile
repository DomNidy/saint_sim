# This image has all necessary tools & libs to compile a go application, but it is pretty big
# so we'll use a multi-stage build step, and copy the compiled binary into a lighter weight base image to run
FROM golang:1.21 AS builder

# This instructs Docker to use this directory as the default destination for all subsequent commands
# create directory inside image that we're building
WORKDIR /app

# Copy the go.mod and go.sum files first to leverage Docker cache for dependencies
COPY go.mod go.sum ./

# Download dependencies (also to use cache) 
RUN go mod download

# Build the saint-sim binary
RUN go build -o saint-sim-bot cmd/sim-bot/main.go
