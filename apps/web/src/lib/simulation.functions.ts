import { createServerFn } from "@tanstack/react-start";
import { z } from "zod";
import { requireAuthMiddleware } from "@/lib/auth/auth.middleware";
import { getSimulation, simulate } from "@/lib/saint-api/generated";
import { readSaintApiErrorMessage } from "@/lib/saint-api/read-saint-api-error-message";
import { zSimulationOptions } from "./saint-api/generated/zod.gen";
import { saintApiClient } from "./saint-api/saint-api-client";

export const submitSimulationRequest = createServerFn({ method: "POST" })
	.middleware([requireAuthMiddleware])
	.inputValidator(zSimulationOptions)
	.handler(async ({ data }) => {
		const response = await simulate({
			client: saintApiClient,
			body: data,
		});

		console.log("response", response);

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
