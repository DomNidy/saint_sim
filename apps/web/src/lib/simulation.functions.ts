import { createServerFn } from "@tanstack/react-start";

import {
	fetchSaintApi,
	readSaintApiError,
	readSaintApiJson,
	type SaintSimulationData,
	type SaintSimulationResponse,
} from "#/lib/saint-api.server";
import {
	simulationRequestSchema,
	simulationResultLookupSchema,
} from "#/lib/simulation.schemas";

export const submitSimulationRequest = createServerFn({ method: "POST" })
	.inputValidator(simulationRequestSchema)
	.handler(async ({ data }) => {
		const response = await fetchSaintApi("/simulate", {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify({
				wow_character: {
					region: data.region,
					realm: data.realm,
					character_name: data.character_name,
				},
			}),
		});

		const payload = await readSaintApiJson<SaintSimulationResponse>(response);

		if (!payload.simulation_request_id) {
			throw new Error("Saint API did not return a simulation request id.");
		}

		return {
			simulationRequestId: payload.simulation_request_id,
		};
	});

export const getSimulationResultByRequestId = createServerFn()
	.inputValidator(simulationResultLookupSchema)
	.handler(async ({ data }) => {
		const response = await fetchSaintApi(
			`/report/request/${encodeURIComponent(data.requestId)}`,
		);

		if (response.status === 202 || response.status === 404) {
			return {
				status: "pending" as const,
			};
		}

		if (!response.ok) {
			throw new Error(await readSaintApiError(response));
		}

		const payload = await response.json();

		return {
			status: "complete" as const,
			result: payload as SaintSimulationData,
		};
	});
