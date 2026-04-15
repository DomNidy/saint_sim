import { createServerFn } from "@tanstack/react-start";
import { z } from "zod";
import { requireAuthMiddleware } from "@/lib/auth/auth.middleware";
import { simulationRequestSchema } from "@/lib/saint-api/contracts";
import type { ErrorResponse } from "@/lib/saint-api/generated";
import { getSimulation, simulate } from "@/lib/saint-api/generated";
import { saintApiClient } from "./saint-api/saint-api-client";

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
			body: data,
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

export const getSimulationResult = createServerFn()
	.inputValidator(z.object({ simulationId: z.uuid() }))
	.handler(async ({ data }) => {
		const response = await getSimulation({
			client: saintApiClient,
			path: { id: data.simulationId },
		});

		if (response.error || !response.data) {
			throw new Error(
				readSaintApiErrorMessage(
					response.error,
					`Unable to retrieve simulation status for sim ${data.simulationId}.`,
				),
			);
		}

		return response.data;
	});
