# todo: extend this image somehow:  https://hub.docker.com/r/simulationcraftorg/simc
# todo: need to install golang in this container
FROM simulationcraftorg/simc:latest AS builder

WORKDIR /build
RUN apk --version

# copy golang from a base image so we don't need to install it in the container
# add it to sytem path
COPY --from=golang:1.21-alpine /usr/local/go/ /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}"

COPY --from=root go.work go.work.sum ./

RUN go work edit -dropuse ./apps/discord_bot
RUN go work edit -dropuse ./apps/api

WORKDIR /build/pkg
COPY --from=pkg . .

WORKDIR /build/apps/simulation_worker

# download modules
COPY go.mod go.sum ./
RUN go mod download

# copy the remaining contents of discord_bot into container
COPY . .

RUN go build -o simulation_worker main.go

# alpine index: https://dl-4.alpinelinux.org/alpine/
FROM alpine:3.20 AS runner

ENV SIMC_BINARY_PATH=./simc

WORKDIR /app

# Copy simc binary from previous stage into image
# the simc binary should be in same dir as the simulation_worker binary
# (the /app directory)
COPY --from=builder /app/SimulationCraft/ .



COPY --from=builder /build/apps/simulation_worker/simulation_worker .
RUN chmod +x simulation_worker

# # Install CA certificates
RUN apk update && apk add ca-certificates libcurl libgcc libstdc++
 
ENTRYPOINT ["./simulation_worker"]
