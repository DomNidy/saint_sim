import { env } from "#/env";

export interface SaintSimulationResponse {
	simulation_request_id?: string;
}

export interface SaintSimulationData {
	id?: number;
	request_id?: string;
	sim_result?: string;
}

interface SaintApiFetchOptions extends Omit<RequestInit, "headers"> {
	headers?: HeadersInit;
	authorization?: string;
}

function getSaintApiConfig() {
	return {
		baseUrl: env.SAINT_API_URL ?? "http://localhost:8000",
		apiKey: env.SAINT_API_KEY,
	};
}

function buildSaintApiHeaders({
	headers,
	authorization,
}: Pick<SaintApiFetchOptions, "headers" | "authorization">): Headers {
	const { apiKey } = getSaintApiConfig();

	if (!apiKey) {
		throw new Error(
			"SAINT_API_KEY is not configured for the TanStack app server runtime.",
		);
	}

	const requestHeaders = new Headers(headers);
	requestHeaders.set("Api-Key", apiKey);

	if (authorization) {
		requestHeaders.set("Authorization", authorization);
	}

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
	{ headers, authorization, ...init }: SaintApiFetchOptions = {},
) {
	const { baseUrl } = getSaintApiConfig();

	const response = await fetch(new URL(path, baseUrl), {
		...init,
		headers: buildSaintApiHeaders({ headers, authorization }),
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
