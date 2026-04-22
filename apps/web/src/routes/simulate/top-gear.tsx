import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
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
					<EquipmentDisplayGroup group={group} key={group.groupLabel} />
				))}
			</div>
		</div>
	);
}
