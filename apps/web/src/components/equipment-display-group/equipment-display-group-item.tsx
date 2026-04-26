import { EquipmentCard } from "@/components/equipment-card";
import type { ParsedEquipmentItem } from "@/lib/equipment/types";

type EquipmentDisplayGroupItemProps = {
	item: ParsedEquipmentItem;
	onClick?: (equipment: ParsedEquipmentItem) => void;

	// show selected state on the item
	isSelected?: boolean;

	// this is not about selection state.
	// this just controls if we should
	// show equipped badge
	isEquipped?: boolean;
};

export const EquipmentDisplayGroupItem = ({
	item,
	onClick,
	isSelected,
	isEquipped,
}: EquipmentDisplayGroupItemProps) => {
	const equipment = item.item;

	return (
		<EquipmentCard
			key={item.selectionId}
			displayName={equipment.display_name}
			itemId={equipment.item_id}
			itemLevel={equipment.item_level}
			isEquipped={isEquipped}
			sourceLabel={equipment.source === "bag" ? "Bag" : "Equipped"}
			enchantId={equipment.enchant_id}
			bonusIds={equipment.bonus_ids}
			gemIds={equipment.gem_ids}
			isSelected={isSelected}
			onClick={() => onClick?.(item)}
		/>
	);
};
