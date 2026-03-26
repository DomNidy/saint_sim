-- name: GetApiKeys :one
SELECT * FROM public.api_keys 
WHERE id = $1 LIMIT 1;

-- name: GetApiKeyServiceName :one
SELECT service_name
FROM public.api_keys
WHERE api_key = $1
LIMIT 1;

-- name: CreateSimulationRequest :exec
INSERT INTO public.simulation_request (id, options)
VALUES ($1, $2);

-- name: GetSimulationRequestOptions :one
SELECT options
FROM public.simulation_request
WHERE id = $1
LIMIT 1;

-- name: GetSimulationData :one
SELECT id, request_id, sim_result
FROM public.simulation_data
WHERE id = $1;

-- name: GetSimulationDataByRequestID :one
SELECT id, request_id, sim_result
FROM public.simulation_data
WHERE request_id = $1
LIMIT 1;

-- name: InsertSimulationData :exec
INSERT INTO public.simulation_data (request_id, sim_result)
VALUES ($1, $2);

-- name: ListenNewSimulationData :exec
LISTEN new_simulation_data;
