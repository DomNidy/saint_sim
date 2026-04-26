import { useMemo } from "react";
import { groupLabelForSlot } from "@/lib/equipment/group";
import type { ParsedEquipmentItem } from "@/lib/equipment/types";
import type { WowCharacter } from "@/lib/saint-api/generated";
import { parseSimcAddonExport } from "@/lib/simulation/parse-addon-export";

type UseParseAddonExportHookType = {
	wowCharacter?: WowCharacter;
	equipmentItems: ParsedEquipmentItem[];

	errorMessage?: string;
};
/**
 * Hook that parses SimulationCraft addon export and returns structured data.
 *
 * In addition to the parsed addon export object, parsed equipment is returned as
 * a flat item list. Each item carries its intended display group so views can
 * group or filter the same item array however they need.
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

	const equipmentItems = useMemo(
		() => buildParsedEquipmentItems(wowCharacter),
		[wowCharacter],
	);

	return {
		wowCharacter,
		equipmentItems,
	};
}

function buildParsedEquipmentItems(
	wowCharacter: WowCharacter | undefined,
): ParsedEquipmentItem[] {
	return [
		...Object.values(wowCharacter?.equipped_items ?? {}),
		...(wowCharacter?.bag_items ?? []),
	].map((item, index) => ({
		selectionId: `addon-export:${index}`,
		groupLabel: groupLabelForSlot(item.slot),
		item,
	}));
}
