## Authentication

**Problem:**

- **`connect ETIMEDOUT ...` error when trying to auth**: Postgres database might not be accessible (Better Auth needs access to it to auth).
- **Solution**: Start postgres database using docker compose. Ensure that it is reachable (set web app env vars for reachability, check better auth client config)

