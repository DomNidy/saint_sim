# simulation_worker

This worker consumes queued simulation requests, runs `simc`, and persists either the result or a terminal error back to the database.
