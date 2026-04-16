FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY --from=root go.mod go.sum ./
RUN go mod download

COPY --from=root . .

RUN go build -o /out/discord_bot ./apps/discord_bot

FROM alpine:3.20 AS runner

WORKDIR /app
RUN apk update && apk add --no-cache ca-certificates

COPY --from=builder /out/discord_bot ./discord_bot
RUN chmod +x discord_bot

ENTRYPOINT ["./discord_bot"]
