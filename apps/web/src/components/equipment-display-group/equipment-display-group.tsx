import { formatGroupLabel } from "@/lib/equipment/group";
import type { ParsedEquipmentItem } from "@/lib/equipment/types";
import { EquipmentDisplayGroupItem } from "./equipment-display-group-item";

type EquipmentDisplayGroupProps = {
	groupLabel: string;
	items: ParsedEquipmentItem[];
	onClickEquipment?: (equipment: ParsedEquipmentItem) => void;
	// called to check selected state, if not provided. if not
	// provided, then selected state is false
	isEquipmentSelected?: (equipment: ParsedEquipmentItem) => boolean;
};

export const EquipmentDisplayGroup = ({
	groupLabel,
	items,
	onClickEquipment,
	isEquipmentSelected,
}: EquipmentDisplayGroupProps) => {
	return (
		<div key={groupLabel} className="space-y-2">
			<h4 className="font-semibold text-lg uppercase tracking-wide">
				{formatGroupLabel(groupLabel)}
			</h4>
			<div className="flex flex-col gap-2">
				{items.map((item) => (
					<EquipmentDisplayGroupItem
						item={item}
						isEquipped={item.item.source === "equipped"}
						key={item.selectionId}
						onClick={onClickEquipment}
						isSelected={isEquipmentSelected?.(item) ?? false}
					/>
				))}
			</div>
		</div>
	);
};
