import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import {
	createFileRoute,
	useHydrated,
	useNavigate,
} from "@tanstack/react-router";
import { LoaderCircle, Sparkles } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { type SubmitHandler, useForm, useWatch } from "react-hook-form";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import {
	Form,
	FormControl,
	FormDescription,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { Textarea } from "@/components/ui/textarea";
import { useParseAddonExport } from "@/hooks/use-parse-addon-export";
import {
	localStorageGet,
	localStorageSet,
	PREV_SIMC_PROFILE_KEY,
} from "@/lib/local-storage";
import {
	type SimulationRequestInput,
	simulationRequestSchema,
} from "@/lib/saint-api/contracts";
import type { AddonExportEquipmentItem } from "@/lib/saint-api/generated";
import { submitSimulationRequest } from "@/lib/simulation.functions";
import { cn } from "@/lib/utils";

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

const orderedSlots = [
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
	"finger",
	"trinket",
	"main_hand",
	"off_hand",
] as const;

export const Route = createFileRoute("/simulate")({
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
	component: SimulationForm,
});

function SimulationForm() {
	const hydrated = useHydrated();
	const navigate = useNavigate();
	const [selectedItems, setSelectedItems] = useState<Set<string>>(new Set());

	const form = useForm<SimulationRequestInput>({
		resolver: zodResolver(simulationRequestSchema),
		defaultValues: {
			simc_addon_export: "",
		},
	});

	const simcExport = useWatch({
		control: form.control,
		name: "simc_addon_export",
		defaultValue: "",
	});

	useEffect(() => {
		if (simcExport !== undefined) {
			setSelectedItems(new Set());
		}
	}, [simcExport]);

	// Use previous simc export as default value
	useEffect(() => {
		if (hydrated) {
			const prevProfile = localStorageGet(PREV_SIMC_PROFILE_KEY);
			if (prevProfile !== null) {
				form.reset({ simc_addon_export: prevProfile });
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
	});

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

	const submitHandler: SubmitHandler<SimulationRequestInput> = (values) => {
		void submitMutation.mutateAsync({ data: values });
		if (hydrated) {
			localStorageSet(PREV_SIMC_PROFILE_KEY, values.simc_addon_export);
		}
	};

	const selectedCount = useMemo(() => selectedItems.size, [selectedItems]);

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
					<Form {...form}>
						<form
							className="flex flex-col gap-5"
							onSubmit={form.handleSubmit(submitHandler)}
						>
							<FormField
								control={form.control}
								name="simc_addon_export"
								render={({ field }) => (
									<FormItem>
										<FormLabel>SimC addon export</FormLabel>
										<FormControl>
											<Textarea
												placeholder={'priest="Example"\nlevel=80\nspec=shadow'}
												autoComplete="off"
												autoCapitalize="none"
												autoCorrect="off"
												className="h-32"
												spellCheck={false}
												{...field}
											/>
										</FormControl>
										<FormDescription>
											Paste the raw output from the SimulationCraft in-game
											addon. Saint will submit it to the backend verbatim.
										</FormDescription>
										<FormMessage />
									</FormItem>
								)}
							/>

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
										form.reset({
											simc_addon_export: "",
										});
									}}
								>
									Clear
								</Button>
							</div>
						</form>
					</Form>

					<div className="space-y-4 border-t pt-4">
						<div className="flex items-center justify-between gap-3">
							<div>
								<h3 className="font-semibold text-lg">Gear preview</h3>
								<p className="text-muted-foreground text-sm">
									Detected items from equipped gear and bags.
								</p>
							</div>
							<Badge variant="secondary">Selected: {selectedCount}</Badge>
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
							<div key={group.slot} className="space-y-2">
								<h4 className="font-medium text-sm uppercase tracking-wide">
									{group.label}
								</h4>
								<div className="grid gap-2 md:grid-cols-2">
									{group.items.map((item) => {
										const isSelected = selectedItems.has(item.fingerprint);

										return (
											<button
												key={item.fingerprint}
												type="button"
												onClick={() => {
													setSelectedItems((current) => {
														const next = new Set(current);
														if (next.has(item.fingerprint)) {
															next.delete(item.fingerprint);
														} else {
															next.add(item.fingerprint);
														}

														return next;
													});
												}}
												className={cn(
													"rounded-md border bg-card p-3 text-left transition-colors",
													isSelected
														? "border-primary ring-1 ring-primary"
														: "hover:border-muted-foreground/40",
												)}
											>
												<div className="flex items-start justify-between gap-3">
													<div className="min-w-0 space-y-1">
														<p className="truncate font-medium text-sm">
															{item.display_name}
														</p>
														<p className="text-muted-foreground text-xs">
															ilvl {item.item_level ?? "?"} ·{" "}
															{item.source === "bag" ? "Bag" : "Equipped"}
														</p>
													</div>
													<a
														href={buildWowheadUrl(item.item_id)}
														data-wowhead={buildWowheadData(item)}
														target="_blank"
														rel="noreferrer"
														onClick={(event) => {
															event.stopPropagation();
														}}
														className="shrink-0 text-xs underline underline-offset-2"
													>
														Wowhead
													</a>
												</div>
											</button>
										);
									})}
								</div>
							</div>
						))}
					</div>
				</CardContent>
			</Card>
		</section>
	);
}

function groupEquipment(items: AddonExportEquipmentItem[]) {
	const groupsBySlot = new Map<string, AddonExportEquipmentItem[]>();

	for (const item of items) {
		const group = groupsBySlot.get(item.slot) ?? [];
		group.push(item);
		groupsBySlot.set(item.slot, group);
	}

	const groups = Array.from(groupsBySlot.entries()).map(
		([slot, groupedItems]) => ({
			slot,
			label: slotLabel(slot),
			items: groupedItems,
		}),
	);

	return groups.sort((left, right) => {
		const leftIndex = orderedSlots.indexOf(
			left.slot as (typeof orderedSlots)[number],
		);
		const rightIndex = orderedSlots.indexOf(
			right.slot as (typeof orderedSlots)[number],
		);

		if (leftIndex === -1 && rightIndex === -1) {
			return left.slot.localeCompare(right.slot);
		}
		if (leftIndex === -1) {
			return 1;
		}
		if (rightIndex === -1) {
			return -1;
		}

		return leftIndex - rightIndex;
	});
}

function slotLabel(slot: string) {
	const words = slot.split("_");
	return words
		.map((word) => word.charAt(0).toUpperCase() + word.slice(1))
		.join(" ");
}

function buildWowheadUrl(itemId: number) {
	return `https://www.wowhead.com/item=${itemId}`;
}

function buildWowheadData(item: AddonExportEquipmentItem) {
	const pairs = [`item=${item.item_id}`];

	if ((item.bonus_ids ?? []).length > 0) {
		pairs.push(`bonus=${(item.bonus_ids ?? []).join(":")}`);
	}
	if (item.enchant_id != null) {
		pairs.push(`ench=${item.enchant_id}`);
	}
	if ((item.gem_ids ?? []).length > 0) {
		pairs.push(`gems=${(item.gem_ids ?? []).join(":")}`);
	}
	if (item.item_level != null) {
		pairs.push(`ilvl=${item.item_level}`);
	}

	return pairs.join("&");
}
