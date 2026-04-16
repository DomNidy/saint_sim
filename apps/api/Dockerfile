FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY --from=root go.mod go.sum ./
RUN go mod download

COPY --from=root . .

RUN go build -o /out/api ./apps/api

FROM alpine:3.20 AS runner

WORKDIR /app
RUN apk update && apk add --no-cache ca-certificates curl

COPY --from=builder /out/api ./api
RUN chmod +x api

ENTRYPOINT ["./api"]
