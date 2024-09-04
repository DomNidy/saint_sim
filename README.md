# saint_sim

The `saint_sim` project aims to provide World of Warcraft players with helpful insights to improve their character's effectiveness and make informed gearing decisions. We provide an interface around the core simulation engine: [simc](https://github.com/simulationcraft/simc), offering an API server and Discord bot.

## Project Structure

_The project structure is subject to change as things are ironed out throughout development._

We use [go workspaces](https://go.dev/doc/tutorial/workspaces) to manage the multiple different modules used in this repository.

- `/apps`: Directory containing the various applications

  - `/apps/discord_bot`: The Discord bot application
  - `/apps/api`: The API server application

- `/pkg`: Directory containing modules shared and used throughout the applications defined in `/apps`

  - `/pkg/secrets`: Utility for reading secrets into memory
  - `/pkg/interfaces`: Contains shared types, interfaces, and models

## Running

We use docker and docker compose to build and deploy the applications. To make the modules inside of `/pkg` available in our dockerfiles, we use the `additional_contexts` argument in our `docker-compose.yml` file.

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

### Issues with go workspaces and dockerfiles

There is some added complexity with our dockerfiles due to the fact that we are also using go workspaces. In each dockerfile, we need to copy the `go.work` and `go.work.sum` files from the root of the repository, into the container. Then, we need to edit the go.work file using `go work edit --dropuse <module_path>` to exclude modules that we don't import in that specific container. If we do not do this, we will run into build errors. As go will try to resolve the paths to those modules, and throw an error (as we don't have them in that container).
