import { zodResolver } from "@hookform/resolvers/zod";
import { createFileRoute } from "@tanstack/react-router";
import { Sparkles } from "lucide-react";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import type z from "zod";
import { AddonExportTextarea } from "@/components/addon-export-textarea";
import { EquipmentDisplayGroup } from "@/components/equipment-display-group/equipment-display-group";
import { SimulationCoreConfigSection } from "@/components/simulation-form/simulation-core-config-section";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";
import { useParseAddonExport } from "@/hooks/use-parse-addon-export";
import { zSimulationConfigTopGear } from "@/lib/saint-api/generated/zod.gen";

export const Route = createFileRoute("/simulate/top-gear")({
	component: RouteComponent,
});

function RouteComponent() {
	const form = useForm<z.infer<typeof zSimulationConfigTopGear>>({
		resolver: zodResolver(zSimulationConfigTopGear),
		defaultValues: {
			kind: "topGear",
		},
		resetOptions: {
			keepErrors: true,
		},
	});

	const [addonExportRaw, setAddonExportRaw] = useState<string>("");
	const parseAddonExportEnabled = !!addonExportRaw && addonExportRaw.length > 0;
	const { equipmentGroups, errorMessage: parseAddonExportError } =
		useParseAddonExport(addonExportRaw, parseAddonExportEnabled);

	const [selectedItems, setSelectedItems] = useState<Set<string>>(new Set());

	useEffect(() => {
		if (equipmentGroups?.length === 0) {
			return;
		}

		window.$WowheadPower?.refreshLinks?.();
	}, [equipmentGroups]);

	return (
		<div>
			<AddonExportTextarea
				value={addonExportRaw}
				onChange={(e) => setAddonExportRaw(e.target.value)}
			/>
			<Form {...form}>
				<form
					className="flex flex-col gap-5"
					onSubmit={(e) => {
						e.preventDefault();
						console.log(form.getValues());
					}}
				>
					<p>Top Gear is a W.I.P</p>

					{form.formState?.errors?.root?.server?.message && (
						<Alert variant={"destructive"}>
							{form.formState?.errors?.root?.server?.message}
						</Alert>
					)}

					<div className="flex flex-wrap items-center gap-3">
						<Button disabled={false} type="submit">
							<Sparkles data-icon="inline-start" />
							Run simulation
						</Button>
						<Button
							type="button"
							variant="secondary"
							onClick={() => {
								form.reset({
									kind: "topGear",
								});
							}}
						>
							Clear
						</Button>
					</div>
				</form>
				<SimulationCoreConfigSection />
			</Form>

			{addonExportRaw.trim().length === 0 ? (
				<p className="text-muted-foreground text-sm">
					Paste a SimC addon export to preview parsed gear.
				</p>
			) : null}

			{parseAddonExportError ? (
				<p className="text-destructive text-sm">
					Error: {parseAddonExportError}
				</p>
			) : null}

			{equipmentGroups && equipmentGroups.length === 0 ? (
				<p className="text-muted-foreground text-sm">
					No gear lines were found in this export.
				</p>
			) : (
				<div className="grid grid-cols-2 gap-2">
					{equipmentGroups?.map((group) => (
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
			)}
		</div>
	);
}
