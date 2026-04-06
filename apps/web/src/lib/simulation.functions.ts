import { createServerFn } from "@tanstack/react-start";
import { z } from "zod";

import { simulationRequestSchema } from "#/lib/saint-api/contracts";
import type { ErrorResponse } from "#/lib/saint-api/generated";
import { getSimulation, simulate } from "#/lib/saint-api/generated";
import { requireAuthMiddleware } from "#/lib/auth.middleware";
import { saintApiClient } from "./saint-api/saint-api-client";



const simulationResultLookupSchema = z.object({
	requestId: z.uuid(),
});

/**
 * Extract API message from response when present, otherwise use the 
 * supplied fallback.
 */
function readSaintApiErrorMessage(
	error: string | ErrorResponse | undefined,
	fallback: string,
) {
	if (error === undefined) {
		return fallback;
	}

	if (typeof error === "string") {
		return error.trim();
	}

	if (error?.message) {
		return error.message;
	}

	return fallback;
}

export const submitSimulationRequest = createServerFn({ method: "POST" })
	.middleware([requireAuthMiddleware])
	.inputValidator(simulationRequestSchema)
	.handler(async ({ data }) => {
		const response = await simulate({
			client: saintApiClient,
			body: {
				wow_character: data,
			},
		});

		if (response.error || !response.data) {
			throw new Error(
				readSaintApiErrorMessage(
					response.error,
					"Unable to submit simulation request.",
				),
			);
		}

		const payload = response.data;

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
		const response = await getSimulation({
			client: saintApiClient,
			path: { id: data.requestId },
		});

		if (response.error || !response.data) {
			throw new Error(
				readSaintApiErrorMessage(
					response.error,
					`Unable to retrieve simulation status for ${data.requestId}.`,
				),
			);
		}

		const payload = response.data;

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
