# saint_sim

The `saint_sim` project aims to provide World of Warcraft players with helpful insights to improve their character's effectiveness and make informed gearing decisions. We provide an interface around the core simulation engine: [simc](https://github.com/simulationcraft/simc), offering an API server and Discord bot.

## Project Structure

_The project structure is subject to change as things are ironed out throughout development._

### A Modular Monolithic Monorepo

The project is structured in such a way that each app in `/app` can be deployed independently as a micro-service, or the entire application can be deployed as a monolith. I made this decision because I want to be able to independently scale the `/apps/simulation_worker`, as this task (WoW character simulations) is fairly computationally intensive, and would benefit from the ability to spin up multiple servers to perform sims (if needed).

If I just wanted to scale the simulation worker independently, I wouln't need `/apps/api`, but I decided to introduce this API instead of sending the simulation requets from `/apps/discord_bot` directly to the rabbitmq broker, as this would make it easier to allow multiple frontends to call the API. With this approach, if you wanted to create a web frontend that performs sims, you need only hit `/apps/api`.

We use [Go Workspaces](https://go.dev/doc/tutorial/workspaces) to allow us to include packages from different go modules in this repository.

- `/apps`: Directory containing the various applications
  - `/apps/discord_bot`: The Discord bot application _(forwards requests to `api`)_
  - `/apps/api`: The API server application
  - `/apps/simulation_worker`: Application which handles simulation requests from users by invoking `simc`, then persists the results to the database
- `/pkg`: Directory containing modules and packages shared throughout `/apps`

  - `/pkg/secrets`: Utility for reading secrets into memory
  - `/pkg/interfaces`: Contains automatically generated, shared types
  - `/pkg/utils`: Miscellanious shared utilities

- `/db`: Contains postgres db initialization scripts, which are copied into the postgres container, then executed. _(Note: these only are executed if the postgres container is started with an empty data directory, read the [image docs](https://hub.docker.com/_/postgres) for more details.)\_

## Environment variables & configuration

The services defined in `docker-compose.yml` depend on environment variables at runtime. You can create a `.env` file in the root directory, and define your secrets in there.

To view an example `.env` file configuration, you can read the `.env.example` file. Generally, if there is a networking related error, the issue will lie with the configuration of these environment variables, or the `docker-compose.yml` configuration.

### `discord_bot` api key

In order for the `discord_bot` to authenticate with the `api`, you will need to generate an API key, hash it with sha256, and then insert this key into the database. You can use the `generate_api_key.sh` script when running locally in order to automatically do this, but you **still need to pass this key as an environment variable to** `discord_bot`. _Note: The database needs to be running in order for it to actually be inserted_

## Management UI's

The ports which management UIs are hosted at can be configured in the `docker-compose.yml` file. The credentials used to login to these UIs should be specified in the `.env` file, colocated in the same directory as the compose file. Here are the default ports:

### pgAdmin - `http://localhost:5050`

Use this to inspect the postgres db.

### rabbitmq management - `http://localhost:15672`

Use this to view the queues.

## Running

Docker and Docker compose are used to build and run the app locally. To make the modules inside of `/pkg` available in our dockerfiles, we use the `additional_contexts` argument in our `docker-compose.yml` file.

The secrets used in `docker-compose.yml` should be stored in a .env file, collocated in the same directory.

**Important**: If you are running the `discord_bot`, ensure you have generated an API key, and inserted this key into the database. This can be done automatically with the `generate_api_key.sh` script.

### To start/stop the all services (containers) locally

```sh
./local.sh start # to stop, pass 'stop' as an argument
```

### Stop a service, rebuild it's image, then start it

```sh
./local.sh api # could also be discord_bot, simulation_worker, etc.
```

### Docker compose

The `local.sh` script simply uses docker compose commands, so you can also interact with the application using the compose cli if you're farmiliar with it.

### Issues with go workspaces and dockerfiles

There is some added complexity with our dockerfiles due to the fact that we are also using go workspaces. In each dockerfile, we need to copy the `go.work` and `go.work.sum` files from the root of the repository, into the container. Then, we need to edit the go.work file using `go work edit --dropuse <module_path>` to exclude modules that we don't import in that specific container. If we do not do this, we will run into build errors. As go will try to resolve the paths to those modules, and throw an error (as we don't have them in that container).
