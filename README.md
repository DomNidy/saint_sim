# saint_sim

The `saint_sim` project aims to provide World of Warcraft players with helpful insights to improve their character's effectiveness and make informed gearing decisions. We provide an interface for the core simulation engine, [simc](https://github.com/simulationcraft/simc), offering an API server and Discord bot.

## Project Structure

_The project structure is subject to change as things are ironed out throughout development._

We use [Go Workspaces](https://go.dev/doc/tutorial/workspaces) to allow us to share packages from different go modules in this repository.

- `/apps`: Directory containing the various applications
  - `/apps/discord_bot`: The Discord bot application _(forwards requests to `api`)_
  - `/apps/api`: The API server application
  - `/apps/simulation_worker`: Application which handles simulation requests from users by invoking `simc`, then persists the results to the database
- `/justfile`: Root task runner for local development and maintenance commands
- `/pkg`: Directory containing shared packages and generated contracts used throughout `/apps`

  - `/pkg/go-shared`: Shared Go workspace modules
  - `/pkg/go-shared/api_types`: Automatically generated Go types from the OpenAPI schema
  - `/pkg/go-shared/auth`: Provides mechanisms for authenticating user requests
  - `/pkg/go-shared/db`: Generated Go database access code from `sqlc`
  - `/pkg/go-shared/secrets`: Utility for reading secrets into memory
  - `/pkg/go-shared/utils`: Miscellaneous shared utilities
  - `/pkg/ts-shared`: Shared generated TypeScript contracts
  - `/pkg/ts-shared/api`: OpenAPI-derived TypeScript types
  - `/pkg/ts-shared/db`: `sqlc`-generated TypeScript query bindings

- `/db/migrations`: Goose SQL migrations. This is the single source of truth for database schema changes.

## Running

> The services defined in `docker-compose.yml` depend on environment variables at runtime. Create `.env` in the repository root with `just setup` before starting the stack. See [Environment Variables & Configuration](#environment-variables--configuration) for the expected values.

Prerequisites:

- [`just`](https://github.com/casey/just)
- [`goose`](https://github.com/pressly/goose)
- Docker
- Go
- A Bash-compatible shell *(WSL is recommended on Windows)*

### Local setup

```sh
just setup
```

The default values in `.env.example` are intended for local Docker development. Update these before starting services:

- `DISCORD_TOKEN` and `APPLICATION_ID` if you want `discord_bot` to connect to Discord.
- `SAINT_API_KEY` after generating and inserting a local API key with `just api-key`.

### Apply database migrations locally

Start Postgres, then run Goose:

```sh
just db-start
just db-migrate
```

Goose migrations in `/db/migrations` are the authoritative schema history for this repository.

### Getting started

```sh
just setup
just db-start
just db-migrate
just api-key
just start
```

Then copy the printed `API key:` value into `SAINT_API_KEY` in `.env`, and recreate the Discord bot:

```sh
just restart discord-bot
```

## Environment Variables & Configuration

The services defined in `docker-compose.yml` depend on environment variables at runtime. Create a `.env` file in the repository root with:

```sh
just setup
```

`.env.example` includes all variables currently referenced by `docker-compose.yml` and the `just` recipes. Generally speaking, if there is a networking or authentication-related error, the issue will lie in the configuration of these environment variables and/or the `docker-compose.yml` configuration itself.

> Docker and Docker compose are used to containerize and run the app locally, and the `docker-compose.yml` file depends on environment variables at runtime.

### Local environment variables

| Variable           | Used for                                                 | Local default / note                                                                       |
| ------------------ | -------------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| `DB_USER`          | Postgres username                                        | `saint`                                                                                    |
| `DB_PASSWORD`      | Postgres password                                        | `saint_dev_password`                                                                       |
| `DB_HOST`          | Postgres hostname used by containers                     | `postgres`                                                                                 |
| `DB_NAME`          | Database name used by apps and scripts                   | `saint`                                                                                    |
| `PGADMIN_EMAIL`    | pgAdmin login email                                      | `admin@example.com`                                                                        |
| `PGADMIN_PASSWORD` | pgAdmin login password                                   | `admin`                                                                                    |
| `RABBITMQ_PORT`    | Host port mapped to RabbitMQ `5672`                      | `5672`                                                                                     |
| `RABBITMQ_USER`    | RabbitMQ username                                        | `saint`                                                                                    |
| `RABBITMQ_PASS`    | RabbitMQ password                                        | `saint_dev_password`                                                                       |
| `SAINT_API_URL`    | Internal API URL used by `discord_bot`                   | `http://api:8080`                                                                          |
| `SAINT_API_KEY`    | API key used by `discord_bot` to authenticate with `api` | Replace after running `just api-key`                                                       |
| `DISCORD_TOKEN`    | Discord bot token                                        | Required to run `discord_bot` against Discord                                              |
| `APPLICATION_ID`   | Discord application ID                                   | Required to run `discord_bot` against Discord                                              |
| `SIMC_IMAGE`       | Base image `simulation_worker` builds off of             | Default will use the latest version. This is a build-time argument, not an actual env var. |

### Authenticating the `discord_bot` with the `api`

`discord_bot` authenticates with the saint API using an API key. If you wish to run the `discord_bot`, you must generate an API key, hash it with sha256, and then insert it into the database. You can use `just api-key` to do this automatically when running locally, but **you still need to update the `.env` file with the printed `SAINT_API_KEY`** so `discord_bot` has access to it at runtime. The database needs to be running in order for the API key to be inserted.

### Changing Postgres or RabbitMQ credentials

This repository persists Postgres and RabbitMQ state in Docker volumes. Because of that, changing `DB_USER`, `DB_PASSWORD`, `RABBITMQ_USER`, or `RABBITMQ_PASS` in `.env` after those services have already initialized can cause future starts to fail.

Local development fix:

If you intentionally changed Postgres or RabbitMQ credentials for local development, remove the existing service data and let the containers initialize again with the new `.env` values:

```sh
just db-reset
```

Important:

- `just db-reset` stops the local stack and removes the `postgres_data` and `rabbitmq_data` volumes.
- This is destructive for local data. Only use it if you are comfortable resetting the local database and broker state.
- After resetting, re-run Goose migrations with `just db-migrate`, then any local bootstrap steps that depend on persisted data, such as `just api-key`, and update `SAINT_API_KEY` in `.env` if needed.

## Management UI's

The ports that management UIs are hosted on can be configured in the `docker-compose.yml` file. The credentials used to login them should be specified in the `.env` file, located in the same directory as the compose file. Here are the default ports:

| Service  | Description                             | URL                      |
| -------- | --------------------------------------- | ------------------------ |
| pgAdmin  | Use this to inspect the Postgres DB.    | `http://localhost:5050`  |
| RabbitMQ | Use this to monitor the message queues. | `http://localhost:15672` |

## Project Architecture & Philosophy

### A Modular Monolithic Monorepo

The project is structured in such a way that each app in `/apps` can be deployed independently as a micro-service, or the entire application can be deployed as a monolith. This philosophy allows `simulation_worker` to be scaled separately. This is beneficial because WoW character simulations are computationally intensive, and would benefit from the ability to spin up multiple servers to perform simulations (if needed).

If I just wanted to scale `simulation_worker` independently, the `api` layer would be unnecessary, as I could send sim requests directly to a message queue and process them asynchronously. However, introducing an API layer and using it as a proxy for sim requests offers many benefits, such as:

- Decoupling the front-end (Discord) from business logic _(which makes it easier to add support for additional front-end clients like web apps)_.

- Promoting better separation of concerns _(which makes development easier as the codebase grows)_

- Making it easier to add additional features *(as a product of more constrained SoC)*

Currently, the front end (Discord) forwards simulation requests from users to a RabbitMQ broker, which then routes the request to a worker (a container running the `/simulation_worker` service). After the simulation is processed, the results are persisted to the database.

## FAQ / Info

### Issues with outdated `simc` version being used

The default env var for `SIMC_IMAGE` uses the `latest` tag. Docker can cache this, and your local `latest` version may be outdated. To solve this and update to the latest image, you can run:

```bash
docker image pull simulationcraftorg/simc:latest
```

### Issues with go workspaces and dockerfiles

There is some added complexity with our dockerfiles since we are also using go workspaces. In each dockerfile, we need to copy the `go.work` and `go.work.sum` files from the root of the repository, into the container. Then, we need to edit the `go.work` file using `go work edit --dropuse <module_path>` to exclude modules that we don't import in that specific container. If we do not do this, we will run into build errors. As go will try to resolve the paths to those modules, then throw an error (because the modules aren't in containers).
