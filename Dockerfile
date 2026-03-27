ARG GO_VERSION=1.21.6
ARG SIMC_IMAGE=simulationcraftorg/simc:latest@sha256:572d6eb923e7fd19bd5007a554ff82bfee9912d591893761844349c67542c28a
ARG SIMC_PLATFORM=linux/amd64

FROM --platform=${SIMC_PLATFORM} ${SIMC_IMAGE} AS simc

FROM golang:${GO_VERSION}-alpine AS deps

RUN apk add --no-cache bash ca-certificates libcurl libgcc libstdc++
# Set environment variables so agent and go code have access to simc
ENV SIMC_HOME=/opt/SimulationCraft
ENV SIMC_BINARY_PATH=/opt/SimulationCraft/simc
ENV PATH=/opt/SimulationCraft:${PATH}

WORKDIR /app

COPY . /app

WORKDIR /app/apps/api
RUN go mod download

WORKDIR /app/apps/simulation_worker
RUN go mod download

WORKDIR /app

# Compile workspace modules explicitly; the repo root is a go.work root, not a module.
RUN go build ./apps/api/... ./apps/simulation_worker/... ./pkg/go-shared/api_types/... ./pkg/go-shared/db/... ./pkg/go-shared/secrets/... ./pkg/go-shared/utils/...
# Compile test binaries without running them, ensuring the Go test toolchain is cached.
RUN go test -run XDUMMY ./apps/api/... ./apps/simulation_worker/... ./pkg/go-shared/api_types/... ./pkg/go-shared/db/... ./pkg/go-shared/secrets/... ./pkg/go-shared/utils/... || true

RUN mkdir -p ${SIMC_HOME}
COPY --from=simc /app/SimulationCraft/simc ${SIMC_HOME}/simc

# COPY docker/run-tests.sh /usr/local/bin/run-tests.sh
# RUN chmod +x /usr/local/bin/run-tests.sh

# CMD ["/usr/local/bin/run-tests.sh"]
