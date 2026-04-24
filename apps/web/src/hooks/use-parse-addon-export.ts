import { useMemo } from "react";
import { groupEquipment } from "@/lib/equipment/group";
import type { EquipmentGroup } from "@/lib/equipment/types";
import type { WowCharacter } from "@/lib/saint-api/generated";
import { parseSimcAddonExport } from "@/lib/simulation/parse-addon-export";

type UseParseAddonExportHookType = {
	wowCharacter?: WowCharacter;
	equipmentGroups?: EquipmentGroup[];

	errorMessage?: string;
};
/**
 * Hook that parses SimulationCraft addon export and returns structured data.
 *
 * In addition to the parsed addon export object, a `equipmentGroups` object is also
 * returned (derived from the parsed items). Each item is grouped by its
 * intended slot (in WoW).
 *
 * Params:
 * - `rawInput`: raw addon export string to parse.
 * - `enabled`: whether parsing should run at all.
 *
 * Behavior:
 * - will not re-parse when disabled
 *
 * Intended usage:
 * - call this where UI needs parsed addon export data
 */
export function useParseAddonExport(
	rawInput: string,
	enabled: boolean,
): UseParseAddonExportHookType {
	const wowCharacter = useMemo(() => {
		if (enabled) {
			return parseSimcAddonExport(rawInput);
		}
	}, [rawInput, enabled]);

	const equipmentGroups = useMemo(
		() =>
			groupEquipment([
				...Object.values(wowCharacter?.equipped_items ?? {}),
				...(wowCharacter?.bag_items ?? []),
			]),
		[wowCharacter?.equipped_items, wowCharacter?.bag_items],
	);

	return {
		wowCharacter,
		equipmentGroups,
	};
}
