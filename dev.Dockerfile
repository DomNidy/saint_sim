# Shared dev image for all Go services.
# We don't COPY any source — it's bind-mounted at runtime.
FROM golang:1.24.0

RUN go install github.com/air-verse/air@v1.61.7

# Needed for your api healthcheck
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

# Bind-mounted .git is owned by the host user; git's safe.directory
# check fails (exit 128) and breaks `go build`. We don't need VCS
# info stamped into dev binaries anyway.
ENV GOFLAGS="-buildvcs=false"

WORKDIR /src
