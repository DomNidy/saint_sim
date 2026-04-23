import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import {
	createFileRoute,
	useHydrated,
	useNavigate,
} from "@tanstack/react-router";
import { LoaderCircle, Sparkles } from "lucide-react";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import type z from "zod";
import { AddonExportTextarea } from "@/components/addon-export-textarea";
import { EquipmentDisplayGroup } from "@/components/equipment-display-group/equipment-display-group";
import { SimulationCoreConfigSection } from "@/components/simulation-form/simulation-core-config-section";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Form } from "@/components/ui/form";
import { useParseAddonExport } from "@/hooks/use-parse-addon-export";
import {
	localStorageGet,
	localStorageSet,
	PREV_SIMC_PROFILE_KEY,
} from "@/lib/local-storage";
import { zSimulationConfigBasic } from "@/lib/saint-api/generated/zod.gen";
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

	const [simcExport, setSimcExport] = useState<string>("");

	// Use previous simc export as default value
	useEffect(() => {
		if (hydrated) {
			const prevProfile = localStorageGet(PREV_SIMC_PROFILE_KEY);
			if (prevProfile !== null) {
				setSimcExport(prevProfile);
			}
		}
	}, [hydrated]);

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

	// ensure image icons re-render
	useEffect(() => {
		if (equipmentGroups?.length === 0) return;
		window.$WowheadPower?.refreshLinks?.();
	}, [equipmentGroups]);

	// autoupdate form state
	useEffect(() => {
		// biome-ignore lint/style/noNonNullAssertion: we want to allow character to be null so we can reset that field and only that field
		form.setValue("character", wowCharacter!);
	}, [wowCharacter, form]);

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
					<AddonExportTextarea
						placeholder={'priest="Example"\nlevel=80\nspec=shadow'}
						autoComplete="off"
						autoCapitalize="none"
						autoCorrect="off"
						className="h-32"
						spellCheck={false}
						value={simcExport}
						onChange={(e) => setSimcExport(e.target.value)}
					/>

					<Form {...form}>
						<form
							className="flex flex-col gap-5"
							onSubmit={(e) => {
								e.preventDefault();
								void submitMutation.mutateAsync({ data: form.getValues() });
								// write to local storage so users can see the profile again
								// whenever they visit the site/refresh
								if (hydrated) {
									localStorageSet(
										PREV_SIMC_PROFILE_KEY,
										form.getValues().simc_addon_export,
									);
								}
							}}
						>
							{form.formState?.errors?.root?.server?.message && (
								<Alert variant={"destructive"}>
									{form.formState?.errors?.root?.server?.message}
								</Alert>
							)}

							{form.formState?.errors && (
								<Alert variant={"destructive"}>
									{JSON.stringify(form.formState.errors)}
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
									onClick={() => form.reset()}
								>
									Clear
								</Button>
							</div>
						</form>
						<SimulationCoreConfigSection />
					</Form>

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
