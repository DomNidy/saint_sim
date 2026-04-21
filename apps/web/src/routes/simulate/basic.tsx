import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import {
	createFileRoute,
	useHydrated,
	useNavigate,
} from "@tanstack/react-router";
import { LoaderCircle } from "lucide-react";
import { useEffect, useMemo } from "react";
import { useForm, useWatch } from "react-hook-form";
import type z from "zod";
import { EquipmentDisplayGroup } from "@/components/equipment-display-group/equipment-display-group";
import { SimulationForm } from "@/components/simulation-form/simulation-form";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { useParseAddonExport } from "@/hooks/use-parse-addon-export";
import { groupEquipment } from "@/lib/equipment/group";
import {
	localStorageGet,
	localStorageSet,
	PREV_SIMC_PROFILE_KEY,
} from "@/lib/local-storage";
import { zSimulationOptionsBasic } from "@/lib/saint-api/generated/zod.gen";
import { submitSimulationRequest } from "@/lib/simulation.functions";

declare global {
	interface Window {
		// Added by the Wowhead tooltip script loaded in the route head.
		$WowheadPower?: {
			refreshLinks?: () => void;
		};
	}
}

// Configure the global Wowhead tooltip script before it loads:
// - `colorLinks: true` colors item links by quality.
// - `iconizeLinks: false` keeps Wowhead from prepending icons to links.
// - `renameLinks: false` keeps Wowhead from rewriting the link text.
const WOWHEAD_CONFIG_SCRIPT =
	"window.whTooltips={colorLinks:true,iconizeLinks:false,renameLinks:false};";

export const Route = createFileRoute("/simulate/basic")({
	head: () => ({
		scripts: [
			{
				children: WOWHEAD_CONFIG_SCRIPT,
			},
			{
				src: "https://wow.zamimg.com/js/tooltips.js",
			},
		],
	}),
	component: SimulationPage,
});

function SimulationPage() {
	const hydrated = useHydrated();
	const navigate = useNavigate();

	const form = useForm<z.infer<typeof zSimulationOptionsBasic>>({
		resolver: zodResolver(zSimulationOptionsBasic),
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
				from: "/simulate",
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

	// auto-parse addon export from form to display gear list
	const parseQuery = useParseAddonExport(simcExport, true);
	const previewGroups = useMemo(
		() => groupEquipment(parseQuery.data?.addon_export.equipment ?? []),
		[parseQuery.data?.addon_export.equipment],
	);

	useEffect(() => {
		if (previewGroups.length === 0 || !hydrated) {
			return;
		}

		window.$WowheadPower?.refreshLinks?.();
	}, [previewGroups, hydrated]);

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
					<SimulationForm
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

						{parseQuery.isLoading ? (
							<div className="flex items-center gap-2 text-muted-foreground text-sm">
								<LoaderCircle className="size-4 animate-spin" />
								Parsing addon export...
							</div>
						) : null}

						{parseQuery.isError ? (
							<p className="text-destructive text-sm">
								{parseQuery.error.message}
							</p>
						) : null}

						{parseQuery.data && previewGroups.length === 0 ? (
							<p className="text-muted-foreground text-sm">
								No gear lines were found in this export.
							</p>
						) : null}

						{previewGroups.map((group) => (
							<EquipmentDisplayGroup group={group} key={group.groupLabel} />
						))}
					</div>
				</CardContent>
			</Card>
		</section>
	);
}
