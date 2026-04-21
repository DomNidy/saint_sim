import type { EquipmentItem, EquipmentSlot } from "../saint-api/generated";
import type { EquipmentGroup } from "./types";

/**
 * Group a flat list of equipment items by their intended slot.
 * All "head" items go into one group, etc.
 *
 * The finger and trinket slots are special cases:
 * - finger1 and finger2 -> go into a single "finger" group
 * - trinket1 and trinket2 -> go into a single "trinket" group
 *
 * @param items items to turn into equipment groups
 * @returns a list of equipment groups for the provided items
 */
export function groupEquipment(items: EquipmentItem[]): EquipmentGroup[] {
	const groupsBySlot = new Map<string, EquipmentItem[]>();

	for (const item of items) {
		const groupLabel = intendedLabelForSlotType(item.slot);
		const group = groupsBySlot.get(groupLabel) ?? [];
		group.push(item);
		groupsBySlot.set(groupLabel, group);
	}

	const groups = Array.from(groupsBySlot.entries()).map(
		([groupLabel, groupedItems]) =>
			({ items: groupedItems, groupLabel: groupLabel }) as EquipmentGroup,
	);

	return groups;
}

function intendedLabelForSlotType(slot: EquipmentSlot): string {
	if (slot.startsWith("trinket")) {
		return "trinket";
	} else if (slot.startsWith("finger")) {
		return "finger";
	}
	return slot.toString();
}

export function formatGroupLabel(groupLabel: string): string {
	const words = groupLabel.split("_");
	return words
		.map((word) => word.charAt(0).toUpperCase() + word.slice(1))
		.join(" ");
}
