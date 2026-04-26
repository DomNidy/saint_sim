import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { LoaderCircle, Sparkles } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
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
import { submitSimulationRequest } from "@/lib/simulation.functions";

export const Route = createFileRoute("/simulate/top-gear")({
	component: RouteComponent,
});

function RouteComponent() {
	const navigate = useNavigate();
	const form = useForm<z.infer<typeof zSimulationConfigTopGear>>({
		resolver: zodResolver(zSimulationConfigTopGear),
		defaultValues: {
			kind: "topGear",
			core_config: {},
			equipment: [],
		},
		resetOptions: {
			keepErrors: true,
		},
	});

	const [addonExportRaw, setAddonExportRaw] = useState<string>("");
	const parseAddonExportEnabled = !!addonExportRaw && addonExportRaw.length > 0;
	const {
		equipmentItems,
		wowCharacter,
		errorMessage: parseAddonExportError,
	} = useParseAddonExport(addonExportRaw, parseAddonExportEnabled);

	const [selectedItems, setSelectedItems] = useState<Set<string>>(new Set());
	const equipmentGroupLabels = useMemo(
		() => Array.from(new Set(equipmentItems.map((item) => item.groupLabel))),
		[equipmentItems],
	);
	const selectedEquipment = useMemo(
		() =>
			equipmentItems
				.filter((item) => selectedItems.has(item.selectionId))
				.map((item) => item.item),
		[equipmentItems, selectedItems],
	);

	const submitMutation = useMutation({
		mutationFn: submitSimulationRequest,
		onSuccess: ({ simulationRequestId }) => {
			navigate({
				from: "/simulate/top-gear",
				to: "/simulation/$simulationId",
				params: {
					simulationId: simulationRequestId,
				},
			});
		},
		onError: ({ message, name }) => {
			form.setError("root.server", {
				message: `${name}: ${message}`,
			});
		},
	});

	useEffect(() => {
		if (equipmentItems.length === 0) {
			return;
		}

		window.$WowheadPower?.refreshLinks?.();
	}, [equipmentItems]);

	useEffect(() => {
		setSelectedItems(
			new Set(
				equipmentItems
					.filter((item) => item.item.source === "equipped")
					.map((item) => item.selectionId),
			),
		);
	}, [equipmentItems]);

	useEffect(() => {
		if (wowCharacter !== undefined) {
			form.setValue("character", wowCharacter);
		}
	}, [wowCharacter, form]);

	useEffect(() => {
		form.setValue("equipment", selectedEquipment);
	}, [selectedEquipment, form]);

	return (
		<div>
			<AddonExportTextarea
				value={addonExportRaw}
				onChange={(e) => setAddonExportRaw(e.target.value)}
			/>
			<Form {...form}>
				<form
					className="flex flex-col gap-5"
					onSubmit={form.handleSubmit((values) => {
						void submitMutation.mutateAsync({ data: values });
					})}
				>
					<p>Top Gear is a W.I.P</p>

					{form.formState?.errors?.root?.server?.message && (
						<Alert variant={"destructive"}>
							{form.formState?.errors?.root?.server?.message}
						</Alert>
					)}

					<div className="flex flex-wrap items-center gap-3">
						<Button disabled={submitMutation.isPending} type="submit">
							{submitMutation.isPending ? (
								<>
									<LoaderCircle
										data-icon="inline-start"
										className="animate-spin"
									/>
									Sending request
								</>
							) : (
								<>
									<Sparkles data-icon="inline-start" />
									Run simulation
								</>
							)}
						</Button>
						<Button
							type="button"
							variant="secondary"
							onClick={() => {
								setAddonExportRaw("");
								setSelectedItems(new Set());
								form.reset({
									kind: "topGear",
									core_config: {},
									equipment: [],
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

			{equipmentItems.length === 0 ? (
				<p className="text-muted-foreground text-sm">
					No gear lines were found in this export.
				</p>
			) : (
				<div className="grid grid-cols-2 gap-2">
					{equipmentGroupLabels.map((groupLabel) => (
						<EquipmentDisplayGroup
							groupLabel={groupLabel}
							items={equipmentItems.filter(
								(item) => item.groupLabel === groupLabel,
							)}
							key={groupLabel}
							onClickEquipment={(eq) => {
								setSelectedItems((currentSelection) => {
									const nextSelection = new Set(currentSelection);
									if (nextSelection.has(eq.selectionId)) {
										nextSelection.delete(eq.selectionId);
									} else {
										nextSelection.add(eq.selectionId);
									}

									return nextSelection;
								});
							}}
							isEquipmentSelected={(eq) => {
								return selectedItems.has(eq.selectionId);
							}}
						/>
					))}
				</div>
			)}
		</div>
	);
}
