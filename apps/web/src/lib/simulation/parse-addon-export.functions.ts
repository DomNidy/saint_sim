import { createServerFn } from "@tanstack/react-start";
import { z } from "zod";
import { parseAddonExport as requestParseAddonExport } from "@/lib/saint-api/generated";
import { readSaintApiErrorMessage } from "@/lib/saint-api/read-saint-api-error-message";
import { saintApiClient } from "@/lib/saint-api/saint-api-client";

export const parseAddonExport = createServerFn({ method: "POST" })
	.inputValidator(
		z.object({
			rawAddonExport: z.string(),
		}),
	)
	.handler(async ({ data }) => {
		const response = await requestParseAddonExport({
			client: saintApiClient,
			body: {
				simc_addon_export: data.rawAddonExport,
			},
		});

		if (response.error || !response.data) {
			throw new Error(
				readSaintApiErrorMessage(
					response.error,
					"Unable to parse SimC addon export.",
				),
			);
		}

		return response.data;
	});
