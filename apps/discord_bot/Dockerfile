# * It is important to know that the names of directories here aren't entirely arbitrary
# * We name them in accordance with the modules defined in our go.work file at the root of our directory
FROM golang:1.21-alpine AS builder

WORKDIR /build

# copy our go.work file from root of our repository
COPY --from=root go.work go.work.sum ./

# remove the unnecssary modules from go workspace file
RUN go work edit -dropuse ./apps/api
RUN go work edit -dropuse ./apps/simulation_worker

# copy the shared packages we need into here
# pkg is an additional_context defined in docker-compose.yml
WORKDIR /build/pkg
COPY --from=pkg . .

# download modules
WORKDIR /build/apps/discord_bot
COPY go.mod go.sum ./
RUN go mod download

# copy the remaining contents of discord_bot into container
COPY . .

RUN go build -o discord_bot main.go

# alpine index: https://dl-4.alpinelinux.org/alpine/
FROM alpine:3.20 AS runner

WORKDIR /app

# Install CA certificates
RUN apk update && apk add ca-certificates

COPY --from=builder /build/apps/discord_bot/discord_bot .

RUN chmod +x discord_bot

ENTRYPOINT ["./discord_bot"]

