import type { EquipmentItem } from "../saint-api/generated";

export function buildWowheadUrl(itemId: number) {
	return `https://www.wowhead.com/item=${itemId}`;
}

export function buildWowheadData(item: EquipmentItem) {
	const pairs = [`item=${item.item_id}`];

	if ((item.bonus_ids ?? []).length > 0) {
		pairs.push(`bonus=${(item.bonus_ids ?? []).join(":")}`);
	}
	if (item.enchant_id != null) {
		pairs.push(`ench=${item.enchant_id}`);
	}
	if ((item.gem_ids ?? []).length > 0) {
		pairs.push(`gems=${(item.gem_ids ?? []).join(":")}`);
	}
	if (item.item_level != null) {
		pairs.push(`ilvl=${item.item_level}`);
	}

	return pairs.join("&");
}
