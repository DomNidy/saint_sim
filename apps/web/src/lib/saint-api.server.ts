import { env } from "#/env";

export interface SaintSimulationResponse {
	simulation_id?: string;
}

export interface SaintSimulationData {
	id?: string;
	simulation_status?: "in_progress" | "in_queue" | "error" | "complete";
	sim_result?: string;
	error_text?: string;
}

function getSaintApiConfig() {
	return {
		baseUrl: env.SAINT_API_URL ?? "http://localhost:8000",
		apiKey: env.SAINT_API_KEY,
	};
}

function buildSaintApiHeaders({ headers }: { headers?: HeadersInit }): Headers {
	const { apiKey } = getSaintApiConfig();

	if (!apiKey) {
		throw new Error(
			"SAINT_API_KEY is not configured for the TanStack app server runtime.",
		);
	}

	const requestHeaders = new Headers(headers);
	requestHeaders.set("Api-Key", apiKey);

	return requestHeaders;
}

async function readApiError(response: Response) {
	try {
		const body = (await response.json()) as {
			error?: string;
			message?: string;
		};

		return (
			body.message ??
			body.error ??
			`Saint API request failed (${response.status})`
		);
	} catch {
		return `Saint API request failed (${response.status})`;
	}
}

export async function fetchSaintApi(
	path: string,
	{
		headers,
		...init
	}: Omit<RequestInit, "headers"> & { headers?: HeadersInit } = {},
) {
	const { baseUrl } = getSaintApiConfig();

	const response = await fetch(new URL(path, baseUrl), {
		...init,
		headers: buildSaintApiHeaders({ headers }),
	});

	return response;
}

export async function readSaintApiJson<T>(response: Response) {
	if (!response.ok) {
		throw new Error(await readApiError(response));
	}

	return (await response.json()) as T;
}

export async function readSaintApiError(response: Response) {
	return readApiError(response);
}
