import type { QueryArrayConfig, QueryArrayResult } from "pg";

interface Client {
    query: (config: QueryArrayConfig) => Promise<QueryArrayResult>;
}

export const getApiKeyQuery = `-- name: GetApiKey :one
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
LIMIT 1`;

export interface GetApiKeyArgs {
    secretHash: string;
}

export interface GetApiKeyRow {
    secretHash: string;
    principalId: string;
    principalType: string;
    userId: string | null;
    serviceId: string | null;
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
        secretHash: row[0],
        principalId: row[1],
        principalType: row[2],
        userId: row[3],
        serviceId: row[4]
    };
}

export const getJwkByIDQuery = `-- name: GetJwkByID :one
SELECT
    id, "publicKey", "privateKey", "createdAt", "expiresAt"
FROM
    public.jwks
WHERE
    id = $1
LIMIT 1`;

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
INSERT INTO public.simulation(kind, sim_config, owner_id)
    VALUES ($1, $2, $3)
RETURNING
    id, owner_id, status, kind, sim_config, sim_result, simc_raw_json2, raw_simc_input, error_text, created_at, started_at, completed_at`;

export interface CreateSimulationArgs {
    kind: string;
    simConfig: any;
    ownerId: string | null;
}

export interface CreateSimulationRow {
    id: string;
    ownerId: string | null;
    status: string;
    kind: string;
    simConfig: any;
    simResult: any | null;
    simcRawJson2: any | null;
    rawSimcInput: string | null;
    errorText: string | null;
    createdAt: Date;
    startedAt: Date | null;
    completedAt: Date | null;
}

export async function createSimulation(client: Client, args: CreateSimulationArgs): Promise<CreateSimulationRow | null> {
    const result = await client.query({
        text: createSimulationQuery,
        values: [args.kind, args.simConfig, args.ownerId],
        rowMode: "array"
    });
    if (result.rows.length !== 1) {
        return null;
    }
    const row = result.rows[0];
    return {
        id: row[0],
        ownerId: row[1],
        status: row[2],
        kind: row[3],
        simConfig: row[4],
        simResult: row[5],
        simcRawJson2: row[6],
        rawSimcInput: row[7],
        errorText: row[8],
        createdAt: row[9],
        startedAt: row[10],
        completedAt: row[11]
    };
}

export const updateSimulationQuery = `-- name: UpdateSimulation :one
UPDATE
    public.simulation
SET
    sim_result = COALESCE($2, sim_result),
    simc_raw_json2 = COALESCE($3, simc_raw_json2),
    error_text = COALESCE($4, error_text),
    started_at = COALESCE($5, started_at),
    completed_at = COALESCE($6, completed_at),
    raw_simc_input = COALESCE($7, raw_simc_input),
    status = COALESCE($8, status)
WHERE
    id = $1
RETURNING
    id, owner_id, status, kind, sim_config, sim_result, simc_raw_json2, raw_simc_input, error_text, created_at, started_at, completed_at`;

export interface UpdateSimulationArgs {
    id: string;
    simResult: any | null;
    simcRawJson2: any | null;
    errorText: string | null;
    startedAt: Date | null;
    completedAt: Date | null;
    rawSimcInput: string | null;
    status: string | null;
}

export interface UpdateSimulationRow {
    id: string;
    ownerId: string | null;
    status: string;
    kind: string;
    simConfig: any;
    simResult: any | null;
    simcRawJson2: any | null;
    rawSimcInput: string | null;
    errorText: string | null;
    createdAt: Date;
    startedAt: Date | null;
    completedAt: Date | null;
}

export async function updateSimulation(client: Client, args: UpdateSimulationArgs): Promise<UpdateSimulationRow | null> {
    const result = await client.query({
        text: updateSimulationQuery,
        values: [args.id, args.simResult, args.simcRawJson2, args.errorText, args.startedAt, args.completedAt, args.rawSimcInput, args.status],
        rowMode: "array"
    });
    if (result.rows.length !== 1) {
        return null;
    }
    const row = result.rows[0];
    return {
        id: row[0],
        ownerId: row[1],
        status: row[2],
        kind: row[3],
        simConfig: row[4],
        simResult: row[5],
        simcRawJson2: row[6],
        rawSimcInput: row[7],
        errorText: row[8],
        createdAt: row[9],
        startedAt: row[10],
        completedAt: row[11]
    };
}

export const getSimulationOptionsQuery = `-- name: GetSimulationOptions :one
SELECT
    sim_config
FROM
    public.simulation
WHERE
    id = $1
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
SELECT
    id, owner_id, status, kind, sim_config, sim_result, simc_raw_json2, raw_simc_input, error_text, created_at, started_at, completed_at
FROM
    public.simulation
WHERE
    id = $1`;

export interface GetSimulationArgs {
    id: string;
}

export interface GetSimulationRow {
    id: string;
    ownerId: string | null;
    status: string;
    kind: string;
    simConfig: any;
    simResult: any | null;
    simcRawJson2: any | null;
    rawSimcInput: string | null;
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
        ownerId: row[1],
        status: row[2],
        kind: row[3],
        simConfig: row[4],
        simResult: row[5],
        simcRawJson2: row[6],
        rawSimcInput: row[7],
        errorText: row[8],
        createdAt: row[9],
        startedAt: row[10],
        completedAt: row[11]
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

