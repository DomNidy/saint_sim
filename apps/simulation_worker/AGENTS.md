# simulation_worker guidance

This service is not a generic Go app in dev. It runs a Go worker against the SimC runtime, so Docker and hot-reload changes need to preserve both the Go toolchain and the SimC image behavior.

## Dev stack shape

- Production image: `apps/simulation_worker/Dockerfile`
- Dev image: `apps/simulation_worker/dev.Dockerfile`
- Dev compose wiring: `docker-compose.dev.yml`
- Hot reload config: `apps/simulation_worker/.air.toml`

In dev:

- the container starts from `SIMC_IMAGE`, not from the shared root `dev.Dockerfile`
- Compose clears the base image entrypoint and runs `air -c apps/simulation_worker/.air.toml`
- the repo is bind-mounted at `/src`
- Air builds `apps/simulation_worker/tmp/main` and watches `apps/simulation_worker` plus `pkg`

## Worker-specific gotchas

- Keep the dev image based on the SimC image. Copying the SimC binary into an unrelated base image can break runtime dependencies even when the file path looks correct.
- `SIMC_BINARY_PATH` is intentionally different in dev and prod:
  - prod uses `./simc` because the runtime `WORKDIR` is `/app`
  - dev uses `/app/SimulationCraft/simc` because the process runs from `/src`
- The dev compose service must clear the inherited SimC entrypoint. Without that, the container can try to run the base image entrypoint instead of `air`.
- The SimC base image does not provide the Go image's normal `GOPATH` / bin `PATH` setup, so the dev Dockerfile installs `air` via `GOBIN=/usr/local/bin`.
- `apps/simulation_worker/tmp/` is a local build-output directory. Treat it as disposable and never rely on tracked artifacts there.

## Change -> Verify

| Change | Minimum verify |
| --- | --- |
| `worker.go`, `simc.go`, or other worker code | `cd apps/simulation_worker && golangci-lint run` then `cd apps/simulation_worker && GOCACHE=/tmp/go-build go build .` |
| Shared code under `pkg/` used by the worker | narrow lint in the changed package, then `cd apps/simulation_worker && GOCACHE=/tmp/go-build go build .` |
| `apps/simulation_worker/.air.toml` | `docker compose -f docker-compose.yml -f docker-compose.dev.yml config --services` and re-read the watcher/build paths in the file |
| `apps/simulation_worker/dev.Dockerfile` or `docker-compose.dev.yml` | `docker compose -f docker-compose.yml -f docker-compose.dev.yml config --services`; if Docker is available, rebuild only `simulation_worker` |
| `apps/simulation_worker/Dockerfile` | targeted Docker build of `simulation_worker` if Docker is available |

## Helpful commands

```bash
just dev
just simc
cd apps/simulation_worker && GOCACHE=/tmp/go-build go build .
```

- `just dev` starts the dev stack with hot reload.
- `just simc` opens a shell in a temporary SimC image with the `simc` binary on `PATH`.
- If a dev-only change seems wrong, compare the dev Dockerfile against the production Dockerfile before assuming the worker should behave like `api` or `discord_bot`.
