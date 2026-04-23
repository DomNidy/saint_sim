import { formatGroupLabel } from "@/lib/equipment/group";
import type { EquipmentGroup } from "@/lib/equipment/types";
import type { EquipmentItem } from "@/lib/saint-api/generated";
import { EquipmentDisplayGroupItem } from "./equipment-display-group-item";

type EquipmentDisplayGroupProps = {
	group: EquipmentGroup;
	onClickEquipment?: (equipment: EquipmentItem) => void;
	// called to check selected state, if not provided. if not
	// provided, then selected state is false
	isEquipmentSelected?: (equipment: EquipmentItem) => boolean;
};

export const EquipmentDisplayGroup = ({
	group,
	onClickEquipment,
	isEquipmentSelected,
}: EquipmentDisplayGroupProps) => {
	return (
		<div key={group.groupLabel} className="space-y-2">
			<h4 className="font-semibold text-lg uppercase tracking-wide">
				{formatGroupLabel(group.groupLabel)}
			</h4>
			<div className="flex flex-col gap-2">
				{group.items.map((item) => (
					<EquipmentDisplayGroupItem
						item={item}
						isEquipped={item.source === "equipped"}
						key={`${item.raw_line}+${item.source}`}
						onClick={onClickEquipment}
						isSelected={isEquipmentSelected?.(item) ?? false}
					/>
				))}
			</div>
		</div>
	);
};
