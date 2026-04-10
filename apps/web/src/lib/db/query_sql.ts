import type { QueryArrayConfig, QueryArrayResult } from "pg";

interface Client {
    query: (config: QueryArrayConfig) => Promise<QueryArrayResult>;
}

export const getApiKeyQuery = `-- name: GetApiKey :one
SELECT id, created_at, updated_at, last_used_at, visible_hint, secret_hash, principal_id FROM public.api_keys
WHERE secret_hash = $1 LIMIT 1`;

export interface GetApiKeyArgs {
    secretHash: string;
}

export interface GetApiKeyRow {
    id: number;
    createdAt: Date;
    updatedAt: Date;
    lastUsedAt: Date | null;
    visibleHint: string;
    secretHash: string;
    principalId: string;
}

export async function getApiKey(client: Client, args: GetApiKeyArgs): Promise<GetApiKeyRow | null> {
    const result = await client.query({
        text: getApiKeyQuery,
        values: [args.secretHash],
        rowMode: "array"
    });
    if (result.rows.length !== 1) {
        return null;
    }
    const row = result.rows[0];
    return {
        id: row[0],
        createdAt: row[1],
        updatedAt: row[2],
        lastUsedAt: row[3],
        visibleHint: row[4],
        secretHash: row[5],
        principalId: row[6]
    };
}

export const getJwkByIDQuery = `-- name: GetJwkByID :one
SELECT id, "publicKey", "privateKey", "createdAt", "expiresAt"
FROM public.jwks
WHERE id = $1 LIMIT 1`;

export interface GetJwkByIDArgs {
    id: string;
}

export interface GetJwkByIDRow {
    id: string;
    publickey: string;
    privatekey: string;
    createdat: Date;
    expiresat: Date | null;
}

export async function getJwkByID(client: Client, args: GetJwkByIDArgs): Promise<GetJwkByIDRow | null> {
    const result = await client.query({
        text: getJwkByIDQuery,
        values: [args.id],
        rowMode: "array"
    });
    if (result.rows.length !== 1) {
        return null;
    }
    const row = result.rows[0];
    return {
        id: row[0],
        publickey: row[1],
        privatekey: row[2],
        createdat: row[3],
        expiresat: row[4]
    };
}

export const createSimulationQuery = `-- name: CreateSimulation :one
INSERT INTO public.simulation (sim_config, owner_id)
VALUES ($1, $2)
RETURNING id, sim_config, sim_result, error_text, created_at, started_at, completed_at, owner_id`;

export interface CreateSimulationArgs {
    simConfig: any;
    ownerId: string | null;
}

export interface CreateSimulationRow {
    id: string;
    simConfig: any;
    simResult: string | null;
    errorText: string | null;
    createdAt: Date;
    startedAt: Date | null;
    completedAt: Date | null;
    ownerId: string | null;
}

export async function createSimulation(client: Client, args: CreateSimulationArgs): Promise<CreateSimulationRow | null> {
    const result = await client.query({
        text: createSimulationQuery,
        values: [args.simConfig, args.ownerId],
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
        completedAt: row[6],
        ownerId: row[7]
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
RETURNING id, sim_config, sim_result, error_text, created_at, started_at, completed_at, owner_id`;

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
    ownerId: string | null;
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
        completedAt: row[6],
        ownerId: row[7]
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
SELECT id, sim_config, sim_result, error_text, created_at, started_at, completed_at, owner_id
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
    ownerId: string | null;
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
        completedAt: row[6],
        ownerId: row[7]
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

