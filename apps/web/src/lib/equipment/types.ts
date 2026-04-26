import type { EquipmentItem } from "../saint-api/generated";

/**
 * Client-side representation of an equipment row parsed from user input.
 *
 * Keep this separate from the generated API `EquipmentItem`: the API item is
 * the payload we send to Saint, while this wrapper carries UI-only metadata
 * needed to render and select parsed rows. That lets duplicate equipment lines
 * stay independently selectable without leaking browser state into the API
 * contract.
 */
export type ParsedEquipmentItem = {
	/**
	 * Stable ID for this parsed row within the current addon export. Used for
	 * React keys and selection state; not sent to the API.
	 */
	selectionId: string;

	/**
	 * Display grouping bucket derived from the item's equipment slot. Normal slots
	 * use their slot name, while interchangeable slots collapse together
	 * (finger1/finger2 -> "finger", trinket1/trinket2 -> "trinket").
	 */
	groupLabel: string;

	item: EquipmentItem;
};
