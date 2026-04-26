import type { ParsedEquipmentItem } from "@/lib/equipment/types";
import { buildWowheadData, buildWowheadUrl } from "@/lib/equipment/wowhead";
import { cn } from "@/lib/utils";

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
		<button
			key={item.selectionId}
			type="button"
			onClick={() => onClick?.(item)}
			className={cn(
				"border bg-card p-2 text-left transition-colors",
				isSelected
					? "border-primary ring-1 ring-primary"
					: "hover:border-muted-foreground/40",
			)}
		>
			<div className="flex items-start justify-between gap-3">
				<div className="min-w-0 space-y-1">
					<p className="truncate font-medium text-sm">
						{equipment.display_name}
					</p>
					<p
						className={cn(
							"text-xs font-bold tracking-tight",
							isEquipped ? "text-wow-green-500" : "text-muted-foreground",
						)}
					>
						ilvl {equipment.item_level ?? "?"} ·{" "}
						{equipment.source === "bag" ? "Bag" : "Equipped"}
					</p>
				</div>
				{/** biome-ignore lint/a11y/useAnchorContent: <just using it for the equipment icon> */}
				<a
					href={buildWowheadUrl(equipment.item_id)}
					data-wowhead={buildWowheadData(equipment)}
					data-wh-icon-size="medium"
					target="_blank"
					rel="noreferrer"
					onClick={(event) => {
						event.stopPropagation();
					}}
					className=""
				></a>
			</div>
		</button>
	);
};
