// Handwritten facade over generated OpenAPI artifacts.
// Keep app imports stable and isolate generator-specific file names/exports here.
import { z } from "zod";

import {
	type SimcAddonExport,
	type Simulation,
	type SimulationOptions,
	SimulationStatus,
} from "@/lib/saint-api/generated";
import { zSimcAddonExport } from "@/lib/saint-api/generated/zod.gen";

export type { SimcAddonExport, Simulation, SimulationOptions };
export { SimulationStatus };

export const simulationRequestSchema = z.object({
	simc_addon_export: zSimcAddonExport,
});

export type SimulationRequestInput = z.infer<typeof simulationRequestSchema>;

export const gearPreviewItemSchema = z.object({
	fingerprint: z.string(),
	slot: z.string(),
	name: z.string(),
	display_name: z.string(),
	item_id: z.number().int(),
	item_level: z.number().int().nullable().optional(),
	icon_url: z.string().nullable().optional(),
	wowhead_url: z.string(),
	wowhead_data: z.string(),
	source: z.enum(["equipped", "bag"]),
	raw_line: z.string(),
});

export const gearPreviewGroupSchema = z.object({
	slot: z.string(),
	label: z.string(),
	items: z.array(gearPreviewItemSchema),
});

export const gearPreviewResponseSchema = z.object({
	groups: z.array(gearPreviewGroupSchema),
});

export type GearPreviewResponse = z.infer<typeof gearPreviewResponseSchema>;
