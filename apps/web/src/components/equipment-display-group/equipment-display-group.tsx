import { formatGroupLabel } from "@/lib/equipment/group";
import type { EquipmentGroup } from "@/lib/equipment/types";
import { EquipmentDisplayGroupItem } from "./equipment-display-group-item";

type EquipmentDisplayGroupProps = {
	group: EquipmentGroup;
};

export const EquipmentDisplayGroup = ({
	group,
}: EquipmentDisplayGroupProps) => {
	return (
		<div key={group.groupLabel} className="space-y-2">
			<h4 className="font-medium text-sm uppercase tracking-wide">
				{formatGroupLabel(group.groupLabel)}
			</h4>
			<div className="grid gap-2 md:grid-cols-2">
				{group.items.map((item) => (
					<EquipmentDisplayGroupItem
						item={item}
						key={`${item.raw_line}+${item.source}`}
					/>
				))}
			</div>
		</div>
	);
};
