ARG SIMC_IMAGE=simulationcraftorg/simc:latest
ARG SIMC_PLATFORM=linux/amd64

FROM --platform=${SIMC_PLATFORM} ${SIMC_IMAGE}

COPY --from=golang:1.24-alpine /usr/local/go/ /usr/local/go

ENV PATH="/usr/local/go/bin:/app/SimulationCraft:${PATH}"
ENV GOFLAGS="-buildvcs=false"
ENV SIMC_BINARY_PATH=/app/SimulationCraft/simc

# The SimC base image does not provide the Go image's GOPATH/bin PATH setup.
# Install air into /usr/local/bin so it is on PATH and Compose can invoke it
# as the container command in the dev override.
ENV GOBIN=/usr/local/bin

RUN go install github.com/air-verse/air@v1.61.7

WORKDIR /src
