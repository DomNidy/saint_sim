import type { SimulationResult } from "../saint-api/generated";

export function isResult<K extends SimulationResult["kind"]>(
	result: SimulationResult | undefined,
	kind: K,
): result is Extract<SimulationResult, { kind: K }> {
	return result?.kind === kind;
}
