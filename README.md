# saint_sim

The `saint_sim` project aims to provide World of Warcraft players with helpful insights to improve their character's effectiveness and make informed gearing decisions. We provide an interface around the core simulation engine: [simc](https://github.com/simulationcraft/simc), offering an API server and Discord bot.

## Project Structure

_The project structure is subject to change as things are ironed out throughout development._

We use [go workspaces](https://go.dev/doc/tutorial/workspaces) to manage the multiple different modules used in this repository.

- `/apps`: Directory containing the various applications
  - `/apps/discord_bot`: The Discord bot application
  - `/apps/api`: The API server application
  - `/apps/simulation_worker`: Application which handles simulation requests from users by invoking `simc`, then persists the results to the database
- `/pkg`: Directory containing modules shared and used throughout the applications defined in `/apps`

  - `/pkg/secrets`: Utility for reading secrets into memory
  - `/pkg/interfaces`: Contains automatically generated, shared interfaces/types
  - `/pkg/utils`: Miscellanious shared utilities

- `/db`: Contains postgres db initialization scripts. These scripts are copied into the postgres container, then executed. (Note: these only are executed if the postgres db container is started with a data directory that is empty.)

## Running

We use docker and docker compose to build and deploy the applications. To make the modules inside of `/pkg` available in our dockerfiles, we use the `additional_contexts` argument in our `docker-compose.yml` file.

The secrets used in `docker-compose.yml` should be stored in a .env file, collocated in the same directory.

To start/stop the containers locally:

```sh
./local.sh start # to stop, pass 'stop' as an argument
```

To build the containers:

```sh
docker compose build
```

To run the containers:

```sh
docker compose up
```

### Connecting to database

The docker compose file starts up a postgres and pgadmin instance. You can connect to the pgadmin instance locally at `localhost:5050`. The login credentials should be specified in the environment variables.

### Issues with go workspaces and dockerfiles

There is some added complexity with our dockerfiles due to the fact that we are also using go workspaces. In each dockerfile, we need to copy the `go.work` and `go.work.sum` files from the root of the repository, into the container. Then, we need to edit the go.work file using `go work edit --dropuse <module_path>` to exclude modules that we don't import in that specific container. If we do not do this, we will run into build errors. As go will try to resolve the paths to those modules, and throw an error (as we don't have them in that container).
