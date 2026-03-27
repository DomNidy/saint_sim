ARG GO_VERSION=1.21.6
ARG SIMC_PLATFORM=linux/amd64
ARG SIMC_REPO=https://github.com/simulationcraft/simc.git
ARG SIMC_REF=29c7f2e7f88358b83e9dcd013d87090bf4278d38
ARG SIMC_THREADS=4

FROM --platform=${SIMC_PLATFORM} golang:${GO_VERSION} AS simc-build

ARG SIMC_REPO
ARG SIMC_REF
ARG SIMC_THREADS

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    clang \
    g++ \
    git \
    libcurl4-openssl-dev \
    llvm \
    make \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src

RUN git init SimulationCraft && \
    cd SimulationCraft && \
    git remote add origin ${SIMC_REPO} && \
    git fetch --depth 1 origin ${SIMC_REF} && \
    git checkout FETCH_HEAD

WORKDIR /src/SimulationCraft

RUN make -C /src/SimulationCraft/engine release CXX=clang++ -j ${SIMC_THREADS} OPTS+="-O0"

FROM --platform=${SIMC_PLATFORM} golang:${GO_VERSION} AS deps

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    git \
    libcurl4 \
    libgcc-s1 \
    libstdc++6 \
    && rm -rf /var/lib/apt/lists/*

ENV SIMC_HOME=/opt/SimulationCraft
ENV SIMC_BINARY_PATH=/opt/SimulationCraft/simc
ENV PATH=/opt/SimulationCraft:${PATH}

WORKDIR /app

COPY . /app

RUN go mod download

RUN go build ./...
RUN go test -run XDUMMY ./... || true

RUN mkdir -p ${SIMC_HOME}
COPY --from=simc-build /src/SimulationCraft/engine/simc ${SIMC_HOME}/simc
COPY --from=simc-build /src/SimulationCraft/profiles ${SIMC_HOME}/profiles

# This bakes the clean pre-edit state into the image.
# After the agent runs, we can extract diffs with: git diff
RUN git config --global user.email "worker@sweagent" && \
    git config --global user.name "worker" && \
    git init && git add -A && git commit -m "pre-edit"

CMD ["/bin/bash"]
