# saint_sim

The `saint_sim` project aims to provide World of Warcraft players with helpful insights to improve their character's effectiveness and make informed gearing decisions. We provide an interface for the core simulation engine, [simc](https://github.com/simulationcraft/simc), offering an API server and Discord bot.

## Project Structure

_The project structure is subject to change as things are ironed out throughout development._

We use [Go Workspaces](https://go.dev/doc/tutorial/workspaces) to allow us to share packages from different go modules in this repository.

- `/apps`: Directory containing the various applications
  - `/apps/discord_bot`: The Discord bot application _(forwards requests to `api`)_
  - `/apps/api`: The API server application
  - `/apps/simulation_worker`: Application which handles simulation requests from users by invoking `simc`, then persists the results to the database
- `/pkg`: Directory containing packages shared and used throughout `/apps`

  - `/pkg/auth`: Provides mechanisms for authenticating user requests
  - `/pkg/secrets`: Utility for reading secrets into memory
  - `/pkg/interfaces`: Contains automatically generated, shared types
  - `/pkg/utils`: Miscellaneous shared utilities

- `/db`: Contains SQL scripts used to initialize Postgres. These are copied into the Postgres container on launch and executed. *(Note: these only are executed if the Postgres container is started with an empty data directory, read the [image docs](https://hub.docker.com/_/postgres) for more details)*

## Running

> The services defined in `docker-compose.yml` depend on environment variables at runtime. You can create a `.env` file in the root directory, and define your secrets in there. See [Environment Variables & Configuration](#environment-variables--configuration) for more details.

### To start/stop all services locally

```sh
./local.sh start # to stop, pass 'stop' as an argument
```

### Stop a service, rebuild it's image, then start it

```sh
./local.sh api # could also be discord_bot, simulation_worker, etc.
```

## Environment Variables & Configuration

The services defined in `docker-compose.yml` depend on environment variables at runtime. You can create a `.env` file in the root directory, and define your secrets in there. To view an example `.env` file configuration, you can read the `.env.example` file. Generally speaking, if there is a networking or authentication-related error, the issue will lie in the configuration of these environment variables and/or the `docker-compose.yml` configuration itself.

> Docker and Docker compose are used to containerize and run the app locally, and the `docker-compose.yml` file depends on environment variables at runtime.

### Authenticating the `discord_bot` with the `api`

`discord_bot` authenticates with the saint API using an API key. If you wish to run the `discord_bot`, you must generate an API key, hash it with sha256, and then insert it into the database. You can use the `generate_api_key.sh` script to do this automatically when running locally, but **you still need to update the `.env` file with the `SAINT_API_KEY`** so `discord_bot` has access to it at runtime. _(Note: The database needs to be running in order for the API key to actually be inserted)_

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

### Issues with go workspaces and dockerfiles

There is some added complexity with our dockerfiles since we are also using go workspaces. In each dockerfile, we need to copy the `go.work` and `go.work.sum` files from the root of the repository, into the container. Then, we need to edit the `go.work` file using `go work edit --dropuse <module_path>` to exclude modules that we don't import in that specific container. If we do not do this, we will run into build errors. As go will try to resolve the paths to those modules, then throw an error (because the modules aren't in containers).
