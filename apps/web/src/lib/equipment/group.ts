import { EquipmentSlot } from "../saint-api/generated";

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

// A top gear simulation payload needs to provide at least one item
// that satisfies each group label.
export const REQUIRED_TOP_GEAR_GROUP_LABELS = new Set(
	[
		EquipmentSlot.HEAD,
		EquipmentSlot.NECK,
		EquipmentSlot.SHOULDER,
		EquipmentSlot.BACK,
		EquipmentSlot.CHEST,
		EquipmentSlot.WRIST,
		EquipmentSlot.HANDS,
		EquipmentSlot.WAIST,
		EquipmentSlot.LEGS,
		EquipmentSlot.FEET,
		EquipmentSlot.FINGER1,
		EquipmentSlot.TRINKET1,
		EquipmentSlot.MAIN_HAND,
	].map(groupLabelForSlot),
);
