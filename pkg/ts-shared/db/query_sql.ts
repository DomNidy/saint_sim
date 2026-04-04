import { QueryArrayConfig, QueryArrayResult } from "pg";

interface Client {
    query: (config: QueryArrayConfig) => Promise<QueryArrayResult>;
}

export const getApiKeyByIdQuery = `-- name: GetApiKeyById :one
SELECT id, api_key, created_at FROM public.api_keys 
WHERE id = $1 LIMIT 1`;

export interface GetApiKeyByIdArgs {
    id: string;
}

export interface GetApiKeyByIdRow {
    id: number;
    apiKey: string;
    createdAt: Date | null;
}

export async function getApiKeyById(client: Client, args: GetApiKeyByIdArgs): Promise<GetApiKeyByIdRow | null> {
    const result = await client.query({
        text: getApiKeyByIdQuery,
        values: [args.id],
        rowMode: "array"
    });
    if (result.rows.length !== 1) {
        return null;
    }
    const row = result.rows[0];
    return {
        id: row[0],
        apiKey: row[1],
        createdAt: row[2]
    };
}

export const getApiKeyQuery = `-- name: GetApiKey :one
SELECT id, api_key, created_at FROM public.api_keys
WHERE api_key = $1 LIMIT 1`;

export interface GetApiKeyArgs {
    apiKey: string;
}

export interface GetApiKeyRow {
    id: number;
    apiKey: string;
    createdAt: Date | null;
}

export async function getApiKey(client: Client, args: GetApiKeyArgs): Promise<GetApiKeyRow | null> {
    const result = await client.query({
        text: getApiKeyQuery,
        values: [args.apiKey],
        rowMode: "array"
    });
    if (result.rows.length !== 1) {
        return null;
    }
    const row = result.rows[0];
    return {
        id: row[0],
        apiKey: row[1],
        createdAt: row[2]
    };
}

export const createSimulationQuery = `-- name: CreateSimulation :one
INSERT INTO public.simulation (sim_config)
VALUES ($1)
RETURNING id, sim_config, sim_result, error_text, created_at, started_at, completed_at`;

export interface CreateSimulationArgs {
    simConfig: any;
}

export interface CreateSimulationRow {
    id: string;
    simConfig: any;
    simResult: string | null;
    errorText: string | null;
    createdAt: Date;
    startedAt: Date | null;
    completedAt: Date | null;
}

export async function createSimulation(client: Client, args: CreateSimulationArgs): Promise<CreateSimulationRow | null> {
    const result = await client.query({
        text: createSimulationQuery,
        values: [args.simConfig],
        rowMode: "array"
    });
    if (result.rows.length !== 1) {
        return null;
    }
    const row = result.rows[0];
    return {
        id: row[0],
        simConfig: row[1],
        simResult: row[2],
        errorText: row[3],
        createdAt: row[4],
        startedAt: row[5],
        completedAt: row[6]
    };
}

export const updateSimulationQuery = `-- name: UpdateSimulation :one
UPDATE public.simulation
SET
    sim_result = COALESCE($1, sim_result),
    error_text = COALESCE($2, error_text),
    started_at = COALESCE($3, started_at),
    completed_at = COALESCE($4, completed_at)
WHERE id = $5
RETURNING id, sim_config, sim_result, error_text, created_at, started_at, completed_at`;

export interface UpdateSimulationArgs {
    simResult: string | null;
    errorText: string | null;
    startedAt: Date | null;
    completedAt: Date | null;
    id: string;
}

export interface UpdateSimulationRow {
    id: string;
    simConfig: any;
    simResult: string | null;
    errorText: string | null;
    createdAt: Date;
    startedAt: Date | null;
    completedAt: Date | null;
}

export async function updateSimulation(client: Client, args: UpdateSimulationArgs): Promise<UpdateSimulationRow | null> {
    const result = await client.query({
        text: updateSimulationQuery,
        values: [args.simResult, args.errorText, args.startedAt, args.completedAt, args.id],
        rowMode: "array"
    });
    if (result.rows.length !== 1) {
        return null;
    }
    const row = result.rows[0];
    return {
        id: row[0],
        simConfig: row[1],
        simResult: row[2],
        errorText: row[3],
        createdAt: row[4],
        startedAt: row[5],
        completedAt: row[6]
    };
}

export const getSimulationOptionsQuery = `-- name: GetSimulationOptions :one
SELECT sim_config 
FROM public.simulation
WHERE id = $1
LIMIT 1`;

export interface GetSimulationOptionsArgs {
    id: string;
}

export interface GetSimulationOptionsRow {
    simConfig: any;
}

export async function getSimulationOptions(client: Client, args: GetSimulationOptionsArgs): Promise<GetSimulationOptionsRow | null> {
    const result = await client.query({
        text: getSimulationOptionsQuery,
        values: [args.id],
        rowMode: "array"
    });
    if (result.rows.length !== 1) {
        return null;
    }
    const row = result.rows[0];
    return {
        simConfig: row[0]
    };
}

export const getSimulationQuery = `-- name: GetSimulation :one
SELECT id, sim_config, sim_result, error_text, created_at, started_at, completed_at
FROM public.simulation
WHERE id = $1`;

export interface GetSimulationArgs {
    id: string;
}

export interface GetSimulationRow {
    id: string;
    simConfig: any;
    simResult: string | null;
    errorText: string | null;
    createdAt: Date;
    startedAt: Date | null;
    completedAt: Date | null;
}

export async function getSimulation(client: Client, args: GetSimulationArgs): Promise<GetSimulationRow | null> {
    const result = await client.query({
        text: getSimulationQuery,
        values: [args.id],
        rowMode: "array"
    });
    if (result.rows.length !== 1) {
        return null;
    }
    const row = result.rows[0];
    return {
        id: row[0],
        simConfig: row[1],
        simResult: row[2],
        errorText: row[3],
        createdAt: row[4],
        startedAt: row[5],
        completedAt: row[6]
    };
}

export const listenNewSimulationDataQuery = `-- name: ListenNewSimulationData :exec
LISTEN new_simulation_data`;

export async function listenNewSimulationData(client: Client): Promise<void> {
    await client.query({
        text: listenNewSimulationDataQuery,
        values: [],
        rowMode: "array"
    });
}

