ARG SIMC_IMAGE=simulationcraftorg/simc:latest
ARG SIMC_PLATFORM=linux/amd64

FROM --platform=${SIMC_PLATFORM} ${SIMC_IMAGE} AS builder

WORKDIR /build
COPY --from=golang:1.24-alpine /usr/local/go/ /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}"

COPY --from=root go.mod go.sum ./
RUN go mod download

COPY --from=root . .

RUN go build -o /out/simulation_worker ./apps/simulation_worker

FROM --platform=${SIMC_PLATFORM} alpine:3.20 AS runner

ENV SIMC_BINARY_PATH=./simc
WORKDIR /app

RUN apk update && apk add --no-cache ca-certificates libcurl libgcc libstdc++

COPY --from=builder /app/SimulationCraft/ .
COPY --from=builder /out/simulation_worker ./simulation_worker
RUN chmod +x simulation_worker

ENTRYPOINT ["./simulation_worker"]
