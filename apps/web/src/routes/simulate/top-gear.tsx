import { createFileRoute } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import type z from "zod";
import { AddonExportTextarea } from "@/components/addon-export-textarea";
import { EquipmentDisplayGroup } from "@/components/equipment-display-group/equipment-display-group";
import { SimulationFormTopGear } from "@/components/simulation-form/simulation-form-topgear";
import { useParseAddonExport } from "@/hooks/use-parse-addon-export";
import type { zSimulationConfigTopGear } from "@/lib/saint-api/generated/zod.gen";

export const Route = createFileRoute("/simulate/top-gear")({
	component: RouteComponent,
});

function RouteComponent() {
	const form = useForm<z.infer<typeof zSimulationConfigTopGear>>({});

	const [addonExportRaw, setAddonExportRaw] = useState<string>("");
	const addonExport = useParseAddonExport(
		addonExportRaw,
		addonExportRaw.length > 0,
	);

	const [selectedItems, setSelectedItems] = useState<Set<string>>(new Set());

	useEffect(() => {
		if (addonExport.data?.groups.length === 0) {
			return;
		}

		window.$WowheadPower?.refreshLinks?.();
	}, [addonExport?.data?.groups]);
	return (
		<div>
			<AddonExportTextarea
				value={addonExportRaw}
				onChange={(e) => setAddonExportRaw(e.target.value)}
			/>
			<SimulationFormTopGear
				form={form}
				isSubmitPending={false}
				submitHandler={(values) => console.log(values)}
			/>
			<div className="grid grid-cols-2 gap-2">
				{addonExport.data?.groups.map((group) => (
					<EquipmentDisplayGroup
						group={group}
						key={group.groupLabel}
						onClickEquipment={(eq) => {
							// todo: we need items to have some actually good id/fingerprint
							// we can't just use raw_line, because it will collide if there
							// are duplicate copies of an item in an addon export. we could
							// derive this server side or transform & derive it on the client.
							if (selectedItems.has(eq.raw_line)) {
								selectedItems.delete(eq.raw_line);
								setSelectedItems(new Set([...selectedItems]));
							} else {
								setSelectedItems(
									new Set([...selectedItems.values(), eq.raw_line]),
								);
							}
						}}
						isEquipmentSelected={(eq) => selectedItems.has(eq.raw_line)}
					/>
				))}
			</div>
		</div>
	);
}
