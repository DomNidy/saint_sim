-- name: GetApiKey :one
SELECT
    api_keys.secret_hash,
    principals.id AS principal_id,
    principals.principal_type,
    principals.user_id,
    principals.service_id
FROM
    public.api_keys
    INNER JOIN public.principals ON principals.id = api_keys.principal_id
WHERE
    secret_hash = $1
LIMIT 1;

-- name: GetJwkByID :one
SELECT
    *
FROM
    public.jwks
WHERE
    id = $1
LIMIT 1;

-- name: CreateSimulation :one
INSERT INTO public.simulation(kind, sim_config, owner_id)
    VALUES ($1, $2, $3)
RETURNING
    *;

-- name: UpdateSimulation :one
UPDATE
    public.simulation
SET
    sim_result = COALESCE(sqlc.narg('sim_result'), sim_result),
    simc_raw_json2 = COALESCE(sqlc.narg('simc_raw_json2'), simc_raw_json2),
    error_text = COALESCE(sqlc.narg('error_text'), error_text),
    started_at = COALESCE(sqlc.narg('started_at'), started_at),
    completed_at = COALESCE(sqlc.narg('completed_at'), completed_at),
    raw_simc_input = COALESCE(sqlc.narg('raw_simc_input'), raw_simc_input),
    status = COALESCE(sqlc.narg('status'), status)
WHERE
    id = $1
RETURNING
    *;

-- name: GetSimulationOptions :one
SELECT
    sim_config
FROM
    public.simulation
WHERE
    id = $1
LIMIT 1;

-- name: GetSimulation :one
SELECT
    *
FROM
    public.simulation
WHERE
    id = $1;

-- name: ListenNewSimulationData :exec
LISTEN new_simulation_data;

