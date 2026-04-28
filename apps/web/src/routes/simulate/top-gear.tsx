import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import {
	createFileRoute,
	useHydrated,
	useNavigate,
} from "@tanstack/react-router";
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
import { REQUIRED_TOP_GEAR_GROUP_LABELS } from "@/lib/equipment/group";
import type { ParsedEquipmentItem } from "@/lib/equipment/types";
import {
	localStorageGet,
	localStorageSet,
	PREV_SIMC_PROFILE_KEY,
} from "@/lib/local-storage";
import { zSimulationConfigTopGear } from "@/lib/saint-api/generated/zod.gen";
import { submitSimulationRequest } from "@/lib/simulation.functions";

export const Route = createFileRoute("/simulate/top-gear")({
	component: RouteComponent,
});

function RouteComponent() {
	const hydrated = useHydrated();
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
				.map((item) => item),
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
		if (hydrated) {
			const prevSimcProfile = localStorageGet(PREV_SIMC_PROFILE_KEY);
			if (prevSimcProfile != null) setAddonExportRaw(prevSimcProfile);
		}
	}, [hydrated]);

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
		form.setValue(
			"equipment",
			selectedEquipment.map((se) => se.item),
		);
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
						if (hydrated) {
							localStorageSet(PREV_SIMC_PROFILE_KEY, addonExportRaw);
						}
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

			<div className="bg-secondary sticky bottom-24">
				<h2 className="font-bold text-2xl">TOP GEAR</h2>
				<p>Selected items: {selectedEquipment.length}</p>
				<p>Total Combinations: {calculateProfilesetCount(selectedEquipment)}</p>
			</div>
		</div>
	);
}

/**
 * Calculate the number of profilesets that can be formed from the provided items.
 * @param items The items that are selected and we want to arrange into profilesets
 * @returns The number of profilesets that could be formed from the selected items
 */
function calculateProfilesetCount(
	selectedItems: ParsedEquipmentItem[],
): number {
	const counts = new Map<string, number>();

	selectedItems.forEach((item) => {
		const cc = counts.get(item.groupLabel);
		if (cc) {
			counts.set(item.groupLabel, cc + 1);
		} else {
			counts.set(item.groupLabel, 1);
		}
	});

	// ensure that we have at least one item of each group label
	// (we cannot form *any* profilesets unless there is at least one)
	for (const requiredGroupLabel of REQUIRED_TOP_GEAR_GROUP_LABELS) {
		const ct = counts.get(requiredGroupLabel);
		if (!ct || ct <= 0) {
			return 0;
		}
	}

	let totalCombinations = 1;

	for (const [itemGroup, count] of counts) {
		if (itemGroup === "finger" || itemGroup === "trinket") {
			totalCombinations *= unorderedPairs(count);
		} else {
			totalCombinations *= count;
		}
	}

	return totalCombinations;
}

/**
 * Calculate the number of unordered pairs that can be formed from
 * a set of n elements
 * @param n The number of elements that we want to compute the number of possible unordered pairs of
 * @returns The number of possible unordered pairs that can be formed from a set of n elements
 */
const unorderedPairs = (n: number): number => {
	if (n < 2) {
		return 0;
	}
	return (n * (n - 1)) / 2;
};
