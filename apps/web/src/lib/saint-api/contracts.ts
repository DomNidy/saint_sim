// Handwritten facade over generated OpenAPI artifacts.
// Keep app imports stable and isolate generator-specific file names/exports here.
import { z } from "zod";

import {
	type AddonExport,
	type ParseAddonExportResponse,
	type SimcAddonExport,
	type Simulation,
	type SimulationOptions,
	SimulationStatus,
} from "@/lib/saint-api/generated";
import { zSimcAddonExport } from "@/lib/saint-api/generated/zod.gen";

export type {
	AddonExport,
	ParseAddonExportResponse,
	SimcAddonExport,
	Simulation,
	SimulationOptions,
};
export { SimulationStatus };

export const simulationRequestSchema = z.object({
	simc_addon_export: zSimcAddonExport,
});

export type SimulationRequestInput = z.infer<typeof simulationRequestSchema>;
