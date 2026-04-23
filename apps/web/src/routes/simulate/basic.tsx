import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import {
	createFileRoute,
	useHydrated,
	useNavigate,
} from "@tanstack/react-router";
import { LoaderCircle } from "lucide-react";
import { useEffect } from "react";
import { useForm, useWatch } from "react-hook-form";
import type z from "zod";
import { EquipmentDisplayGroup } from "@/components/equipment-display-group/equipment-display-group";
import { SimulationFormBasic } from "@/components/simulation-form/simulation-form-basic";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { useParseAddonExport } from "@/hooks/use-parse-addon-export";
import {
	localStorageGet,
	localStorageSet,
	PREV_SIMC_PROFILE_KEY,
} from "@/lib/local-storage";
import {
	zSimcAddonExport,
	zSimulationConfigBasic,
} from "@/lib/saint-api/generated/zod.gen";
import { submitSimulationRequest } from "@/lib/simulation.functions";

export const Route = createFileRoute("/simulate/basic")({
	component: SimulationPage,
});

function SimulationPage() {
	const hydrated = useHydrated();
	const navigate = useNavigate();

	const form = useForm<z.infer<typeof zSimulationConfigBasic>>({
		resolver: zodResolver(zSimulationConfigBasic),
		defaultValues: {
			kind: "basic",
			simc_addon_export: "",
		},
		resetOptions: {
			keepErrors: true,
		},
	});

	const simcExport = useWatch({
		control: form.control,
		name: "simc_addon_export",
		defaultValue: "",
	});

	// Use previous simc export as default value
	useEffect(() => {
		if (hydrated) {
			const prevProfile = localStorageGet(PREV_SIMC_PROFILE_KEY);
			if (prevProfile !== null) {
				form.reset({ kind: "basic", simc_addon_export: prevProfile });
			}
		}
	}, [hydrated, form]);

	const submitMutation = useMutation({
		mutationFn: submitSimulationRequest,
		onSuccess: ({ simulationRequestId }) => {
			navigate({
				from: "/simulate/",
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

	const parseAddonExportEnabled = !!simcExport && simcExport.length > 0; // only parse addon export when we have it
	const {
		equipmentGroups,
		wowCharacter,
		errorMessage: parseAddonExportError,
	} = useParseAddonExport(simcExport, parseAddonExportEnabled);

	useEffect(() => {
		if (equipmentGroups?.length === 0) {
			return;
		}

		window.$WowheadPower?.refreshLinks?.();
	}, [equipmentGroups]);

	return (
		<section className="w-full pb-10 pt-12">
			<Card className="relative overflow-hidden">
				<CardHeader className="gap-2">
					<CardTitle>Submit a Saint simulation</CardTitle>
					<CardDescription>
						Paste your SimulationCraft addon export to submit a simulation.
					</CardDescription>
				</CardHeader>
				<CardContent className="flex flex-col gap-6">
					<SimulationFormBasic
						form={form}
						isSubmitPending={submitMutation.isPending}
						submitHandler={(values) => {
							void submitMutation.mutateAsync({ data: values });
							// write to local storage so users can see the profile again
							// whenever they visit the site/refresh
							if (hydrated) {
								localStorageSet(
									PREV_SIMC_PROFILE_KEY,
									values.simc_addon_export,
								);
							}
						}}
					/>

					{/* Display list of all parsed gear */}
					<div className="space-y-4 border-t pt-4">
						<div className="flex items-center justify-between gap-3">
							<div>
								<h3 className="font-semibold text-lg">Gear preview</h3>
								<p className="text-muted-foreground text-sm">
									Detected items from equipped gear and bags.
								</p>
							</div>
						</div>

						{simcExport.trim().length === 0 ? (
							<p className="text-muted-foreground text-sm">
								Paste a SimC addon export to preview parsed gear.
							</p>
						) : null}

						{parseAddonExportError ? (
							<p className="text-destructive text-sm">
								Error: {parseAddonExportError}
							</p>
						) : null}

						{equipmentGroups && equipmentGroups?.length === 0 ? (
							<p className="text-muted-foreground text-sm">
								No gear lines were found in this export.
							</p>
						) : (
							<div className="grid grid-cols-2 gap-2">
								{equipmentGroups?.map((group) => (
									<EquipmentDisplayGroup group={group} key={group.groupLabel} />
								))}
							</div>
						)}
					</div>
				</CardContent>
			</Card>
		</section>
	);
}
