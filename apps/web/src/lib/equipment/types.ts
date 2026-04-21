import type { EquipmentItem } from "../saint-api/generated";

export type EquipmentGroup = {
	// display label for the group.
	groupLabel: string;

	items: EquipmentItem[];
};
