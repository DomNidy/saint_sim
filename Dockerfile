ARG GO_VERSION=1.21.6
ARG SIMC_IMAGE=simulationcraftorg/simc:latest@sha256:572d6eb923e7fd19bd5007a554ff82bfee9912d591893761844349c67542c28a
ARG SIMC_PLATFORM=linux/amd64

FROM --platform=${SIMC_PLATFORM} ${SIMC_IMAGE} AS simc

FROM golang:${GO_VERSION}-alpine AS deps

WORKDIR /src

COPY go.work go.work.sum ./
COPY apps/api/go.mod apps/api/go.sum ./apps/api/
COPY apps/simulation_worker/go.mod apps/simulation_worker/go.sum ./apps/simulation_worker/
COPY pkg/go-shared ./pkg/go-shared

WORKDIR /src/apps/api
RUN go mod download

WORKDIR /src/apps/simulation_worker
RUN go mod download

FROM golang:${GO_VERSION}-alpine AS test-runner

ENV SIMC_HOME=/opt/SimulationCraft
ENV SIMC_BINARY_PATH=/opt/SimulationCraft/simc
ENV PATH=/opt/SimulationCraft:${PATH}

WORKDIR /src

RUN apk add --no-cache bash ca-certificates libcurl libgcc libstdc++

COPY --from=simc /app/SimulationCraft ${SIMC_HOME}
COPY --from=deps /go/pkg /go/pkg
COPY . .
COPY docker/run-tests.sh /usr/local/bin/run-tests.sh

RUN chmod +x /usr/local/bin/run-tests.sh

CMD ["/usr/local/bin/run-tests.sh"]
