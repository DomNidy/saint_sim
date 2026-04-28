export function buildWowheadUrl(itemId: number) {
	return `https://www.wowhead.com/item=${itemId}`;
}

declare global {
	interface Window {
		// Added by the Wowhead tooltip script added in root route
		$WowheadPower?: {
			refreshLinks?: () => void;
		};
	}
}

type WowheadItemData = {
	item_id: number;
	bonus_ids?: number[];
	enchant_id?: number | null;
	gem_ids?: number[];
	item_level?: number | null;
};

export function buildWowheadData(item: WowheadItemData) {
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
