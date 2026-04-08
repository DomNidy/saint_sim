// Handwritten facade over generated OpenAPI artifacts.
// Keep app imports stable and isolate generator-specific file names/exports here.
import type { z } from "zod";

import {
	type Simulation,
	type SimulationOptions,
	SimulationStatus,
	type WowCharacter,
	WowRealm,
	WowRegion,
} from "@/lib/saint-api/generated";
import { zWowCharacter } from "@/lib/saint-api/generated/zod.gen";

export type { Simulation, SimulationOptions, WowCharacter };
export { SimulationStatus, WowRealm, WowRegion };

export const simulationRegions = Object.values(WowRegion);
export const simulationRealms = Object.values(WowRealm);

export const simulationRequestSchema = zWowCharacter;

export type SimulationRequestInput = z.infer<typeof simulationRequestSchema>;
