FROM golang:1.21-alpine AS builder

WORKDIR /build

COPY --from=root go.work go.work.sum ./

RUN go work edit -dropuse ./apps/discord_bot

WORKDIR /build/pkg
COPY --from=pkg . .

WORKDIR /build/apps/api


# download modules
COPY go.mod go.sum ./
RUN go mod download

# copy the remaining contents of discord_bot into container
COPY . .

RUN go build -o api main.go

# alpine index: https://dl-4.alpinelinux.org/alpine/
FROM alpine:3.20 AS runner

WORKDIR /app

COPY --from=builder /build/apps/api/api .

RUN chmod +x api

# Install CA certificates
RUN apk update && apk add ca-certificates

ENTRYPOINT ["./api"]
