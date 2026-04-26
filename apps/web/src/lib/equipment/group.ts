import type { EquipmentSlot } from "../saint-api/generated";

export function groupLabelForSlot(slot: EquipmentSlot): string {
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
