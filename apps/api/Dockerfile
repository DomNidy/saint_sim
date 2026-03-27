FROM golang:1.21-alpine AS builder

WORKDIR /build

COPY --from=root go.work go.work.sum ./

RUN go work edit -dropuse ./apps/simulation_worker

WORKDIR /build/pkg
COPY --from=pkg . .

WORKDIR /build/apps/api


# download modules
COPY go.mod go.sum ./
RUN go mod download

# copy the remaining contents into container
COPY . .

RUN go build -o api .

# alpine index: https://dl-4.alpinelinux.org/alpine/
FROM alpine:3.20 AS runner

WORKDIR /app

# Install CA certificates
RUN apk update && apk add ca-certificates && apk add curl

COPY --from=builder /build/apps/api/api .

RUN chmod +x api


ENTRYPOINT ["./api"]
