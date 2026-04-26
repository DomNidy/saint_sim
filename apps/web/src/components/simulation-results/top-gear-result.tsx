import { useEffect } from "react";
import { EquipmentCard } from "@/components/equipment-card";
import { formatGroupLabel } from "@/lib/equipment/group";
import type {
	EquipmentItem,
	SimulationResultTopGear,
	TopGearProfilesetItems,
	TopGearProfilesetResult,
} from "@/lib/saint-api/generated";

const TOP_GEAR_SLOT_ORDER = [
	"head",
	"neck",
	"shoulder",
	"back",
	"chest",
	"wrist",
	"hands",
	"waist",
	"legs",
	"feet",
	"finger1",
	"finger2",
	"trinket1",
	"trinket2",
	"main_hand",
	"off_hand",
] as const satisfies readonly (keyof TopGearProfilesetItems)[];

const numberFormatter = new Intl.NumberFormat();
const metricFormatter = new Intl.NumberFormat(undefined, {
	maximumFractionDigits: 2,
});

export function TopGearSimulationResultDisplay(
	result: SimulationResultTopGear,
) {
	useEffect(() => {
		if (result.profilesets.length === 0) {
			return;
		}

		window.$WowheadPower?.refreshLinks?.();
	}, [result.profilesets]);

	return (
		<section className="space-y-4">
			<div className="space-y-1">
				<h2 className="font-semibold text-xl">Top gear results</h2>
				<p className="text-muted-foreground text-sm">
					{result.profilesets.length} profilesets ranked by {result.metric}
				</p>
			</div>

			{result.profilesets.map((profileset) => (
				<TopGearSimulationResultProfileset
					key={profileset.name}
					profileset={profileset}
					equipment={result.equipment}
					metric={result.metric}
				/>
			))}
		</section>
	);
}

type TopGearSimulationResultProfilesetProps = {
	profileset: TopGearProfilesetResult;
	equipment: EquipmentItem[];
	metric: string;
};

function TopGearSimulationResultProfileset({
	profileset,
	equipment,
	metric,
}: TopGearSimulationResultProfilesetProps) {
	return (
		<article className="space-y-4 border bg-card p-4">
			<div className="flex flex-wrap items-start justify-between gap-3">
				<div className="min-w-0">
					<h3 className="truncate font-semibold text-lg">{profileset.name}</h3>
					<p className="text-muted-foreground text-sm">
						Mean {metric}: {numberFormatter.format(profileset.mean)}
					</p>
				</div>

				<div className="text-right">
					<p className="font-semibold text-2xl">
						{metricFormatter.format(profileset.mean)}
					</p>
					<p className="text-muted-foreground text-xs">
						Mean error:{" "}
						{profileset.mean_error == null
							? "unavailable"
							: metricFormatter.format(profileset.mean_error)}
					</p>
				</div>
			</div>

			<div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
				{TOP_GEAR_SLOT_ORDER.map((slot) => {
					const equipmentIndex = profileset.items[slot];

					if (equipmentIndex == null) {
						return null;
					}

					const item = equipment[equipmentIndex];

					return (
						<div key={slot} className="space-y-1.5">
							<p className="font-semibold text-muted-foreground text-xs uppercase">
								{formatGroupLabel(slot)}
							</p>
							{item ? (
								<EquipmentCard
									displayName={item.display_name}
									itemId={item.item_id}
									itemLevel={item.item_level}
									isEquipped={item.source === "equipped"}
									sourceLabel={item.source === "bag" ? "Bag" : "Equipped"}
									enchantId={item.enchant_id}
									bonusIds={item.bonus_ids}
									gemIds={item.gem_ids}
								/>
							) : (
								<div className="border border-destructive/50 bg-destructive/10 p-2 text-destructive text-sm">
									Missing equipment index {equipmentIndex}
								</div>
							)}
						</div>
					);
				})}
			</div>
		</article>
	);
}
