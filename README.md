# saint_sim

The `saint_sim` project aims to provide World of Warcraft players with helpful insights to improve their character's effectiveness and make informed gearing decisions. We provide an interface for the core simulation engine, [simc](https://github.com/simulationcraft/simc), offering an API server.

## FAQ / Info

### Issues with outdated `simc` version being used

The default env var for `SIMC_IMAGE` uses the `latest` tag. Docker can cache this, and your local `latest` version may be outdated. To solve this and update to the latest image, you can run:

```bash
docker image pull simulationcraftorg/simc:latest
```