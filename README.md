# saint-sim

The `saint-sim` project aims to provide World of Warcraft players with helpful insights to improve their character's effectiveness and make informed gearing decisions. We provide an interface around the core simulation engine: [simc](https://github.com/simulationcraft/simc), offering an API server and Discord bot.

## Project Structure

*The project structure is subject to change as things are ironed out throughout development.*

- `/cmd`: The common directory containing the various application entry-points and other interfaces
  - `/cmd/sim-bot`: The Discord bot application
  - `/cmd/sim-api`: The API server application

- `/internal`: Directory containing packages shared and used throughout the applications defined in `/cmd`
  - `/internal/secrets`: For reading secrets into memory