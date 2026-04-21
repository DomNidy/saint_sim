import { buildWowheadData, buildWowheadUrl } from "@/lib/equipment/wowhead";
import type { EquipmentItem } from "@/lib/saint-api/generated";
import { cn } from "@/lib/utils";

type EquipmentDisplayGroupItemProps = {
	item: EquipmentItem;
	onClick?: () => void;

	// show selected state on the item
	isSelected?: boolean;
};

export const EquipmentDisplayGroupItem = ({
	item,
	onClick,
	isSelected,
}: EquipmentDisplayGroupItemProps) => {
	const fp = `${item.raw_line}`;

	return (
		<button
			key={fp}
			type="button"
			onClick={onClick}
			className={cn(
				"rounded-md border bg-card p-3 text-left transition-colors",
				isSelected
					? "border-primary ring-1 ring-primary"
					: "hover:border-muted-foreground/40",
			)}
		>
			<div className="flex items-start justify-between gap-3">
				<div className="min-w-0 space-y-1">
					<p className="truncate font-medium text-sm">{item.display_name}</p>
					<p className="text-muted-foreground text-xs">
						ilvl {item.item_level ?? "?"} ·{" "}
						{item.source === "bag" ? "Bag" : "Equipped"}
					</p>
				</div>
				<a
					href={buildWowheadUrl(item.item_id)}
					data-wowhead={buildWowheadData(item)}
					target="_blank"
					rel="noreferrer"
					onClick={(event) => {
						event.stopPropagation();
					}}
					className="shrink-0 text-xs underline underline-offset-2"
				>
					Wowhead
				</a>
			</div>
		</button>
	);
};
