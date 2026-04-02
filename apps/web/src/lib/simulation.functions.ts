import { createServerFn } from "@tanstack/react-start";

import { requireAuthMiddleware } from "#/lib/auth.middleware";
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
	.middleware([requireAuthMiddleware])
	.inputValidator(simulationRequestSchema)
	.handler(async ({ data }) => {
		const response = await fetchSaintApi("/simulation", {
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

		if (!payload.simulation_id) {
			throw new Error("Saint API did not return a simulation id.");
		}

		return {
			simulationRequestId: payload.simulation_id,
		};
	});

export const getSimulationResultByRequestId = createServerFn()
	.middleware([requireAuthMiddleware])
	.inputValidator(simulationResultLookupSchema)
	.handler(async ({ data }) => {
		const response = await fetchSaintApi(
			`/simulation/${encodeURIComponent(data.requestId)}`,
		);

		if (!response.ok) {
			throw new Error(await readSaintApiError(response));
		}

		const payload = await readSaintApiJson<SaintSimulationData>(response);

		switch (payload.simulation_status) {
			case "complete":
				return {
					status: "complete" as const,
					result: payload,
				};
			case "error":
				return {
					status: "error" as const,
					result: payload,
				};
			case "in_progress":
			case "in_queue":
				return {
					status: "pending" as const,
					result: payload,
				};
			default:
				throw new Error("Saint API returned an unknown simulation status.");
		}
	});
