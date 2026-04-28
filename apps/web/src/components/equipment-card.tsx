import { buildWowheadData, buildWowheadUrl } from "@/lib/equipment/wowhead";
import { cn } from "@/lib/utils";

type EquipmentCardProps = {
	displayName: string;
	itemId: number;
	itemLevel?: number | null;
	isEquipped?: boolean;
	sourceLabel?: string;
	enchantId?: number | null;
	bonusIds?: number[];
	gemIds?: number[];
	isSelected?: boolean;
	onClick?: () => void;
	className?: string;
};

/**
 * Displays data on a piece of WoW equipment in a card, including the item icon & tooltip.
 *
 * NOTE: Requires the dom to have wowhead tooltips.js installed, and when changes occur,
 * the `window.$WowheadPower.refreshLinks() script must be called.
 */
export const EquipmentCard = ({
	displayName,
	itemId,
	itemLevel,
	isEquipped,
	sourceLabel,
	enchantId,
	bonusIds,
	gemIds,
	isSelected,
	onClick,
	className,
}: EquipmentCardProps) => {
	const content = (
		<div className="flex items-start justify-between gap-3">
			<div className="min-w-0 space-y-1">
				<p className="truncate font-medium text-sm">{displayName}</p>
				<p
					className={cn(
						"text-xs font-bold tracking-tight",
						isEquipped ? "text-wow-green-500" : "text-muted-foreground",
					)}
				>
					ilvl {itemLevel ?? "?"}
					{sourceLabel ? ` · ${sourceLabel}` : null}
				</p>
			</div>
			{/** biome-ignore lint/a11y/useAnchorContent: Wowhead injects the item icon into this anchor. */}
			<a
				href={buildWowheadUrl(itemId)}
				data-wowhead={buildWowheadData({
					item_id: itemId,
					bonus_ids: bonusIds,
					enchant_id: enchantId,
					gem_ids: gemIds,
					item_level: itemLevel,
				})}
				data-wh-icon-size="medium"
				target="_blank"
				rel="noreferrer"
				onClick={(event) => {
					event.stopPropagation();
				}}
			></a>
		</div>
	);

	const cardClassName = cn(
		"border bg-card p-2 text-left transition-colors",
		isSelected
			? "border-primary ring-1 ring-primary"
			: "hover:border-muted-foreground/40",
		className,
	);

	if (onClick) {
		return (
			<button type="button" onClick={onClick} className={cardClassName}>
				{content}
			</button>
		);
	}

	return <div className={cardClassName}>{content}</div>;
};
