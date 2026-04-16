import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery } from "@tanstack/react-query";
import {
	createFileRoute,
	useHydrated,
	useNavigate,
} from "@tanstack/react-router";
import { LoaderCircle, Sparkles } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { type SubmitHandler, useForm } from "react-hook-form";
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
import {
	localStorageGet,
	localStorageSet,
	PREV_SIMC_PROFILE_KEY,
} from "@/lib/local-storage";
import {
	type SimulationRequestInput,
	simulationRequestSchema,
} from "@/lib/saint-api/contracts";
import {
	getGearPreview,
	submitSimulationRequest,
} from "@/lib/simulation.functions";
import { cn } from "@/lib/utils";

declare global {
	interface Window {
		$WowheadPower?: {
			refreshLinks?: () => void;
		};
	}
}

const WOWHEAD_CONFIG_SCRIPT =
	"window.whTooltips={colorLinks:true,iconizeLinks:false,renameLinks:false};";

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
	const [debouncedProfile, setDebouncedProfile] = useState("");

	const form = useForm<SimulationRequestInput>({
		resolver: zodResolver(simulationRequestSchema),
		defaultValues: {
			simc_addon_export: "",
		},
	});

	const simcExport = form.watch("simc_addon_export");

	useEffect(() => {
		const timeoutID = setTimeout(() => {
			setDebouncedProfile(simcExport.trim());
			setSelectedItems(new Set());
		}, 450);

		return () => {
			clearTimeout(timeoutID);
		};
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
			// On submit, redirect to the status page for the sim
			navigate({
				from: "/simulate",
				to: "/simulation/$simulationId",
				params: {
					simulationId: simulationRequestId,
				},
			});
		},
	});

	const previewQuery = useQuery({
		queryKey: ["simc-gear-preview", debouncedProfile],
		queryFn: () =>
			getGearPreview({ data: { simc_addon_export: debouncedProfile } }),
		enabled: debouncedProfile.length > 0,
		retry: false,
	});

	useEffect(() => {
		if (!previewQuery.data || !hydrated) {
			return;
		}

		window.$WowheadPower?.refreshLinks?.();
	}, [previewQuery.data, hydrated]);

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

						{debouncedProfile.length === 0 ? (
							<p className="text-muted-foreground text-sm">
								Paste a SimC addon export to preview parsed gear.
							</p>
						) : null}

						{previewQuery.isLoading ? (
							<div className="flex items-center gap-2 text-muted-foreground text-sm">
								<LoaderCircle className="size-4 animate-spin" />
								Building gear preview...
							</div>
						) : null}

						{previewQuery.isError ? (
							<p className="text-destructive text-sm">
								{previewQuery.error.message === "Authentication required."
									? "Sign in to view gear previews."
									: previewQuery.error.message}
							</p>
						) : null}

						{previewQuery.data?.groups.length === 0 ? (
							<p className="text-muted-foreground text-sm">
								No gear lines were found in this export.
							</p>
						) : null}

						{previewQuery.data?.groups.map((group) => (
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
													"flex items-center gap-3 rounded-md border bg-card p-3 text-left transition-colors",
													isSelected
														? "border-primary ring-1 ring-primary"
														: "hover:border-muted-foreground/40",
												)}
											>
												<a
													href={item.wowhead_url}
													data-wowhead={item.wowhead_data}
													target="_blank"
													rel="noreferrer"
													onClick={(event) => {
														event.stopPropagation();
													}}
													className="shrink-0"
												>
													<img
														src={item.icon_url ?? ""}
														alt={item.display_name}
														className="size-12 rounded border"
													/>
												</a>
												<div className="min-w-0 space-y-1">
													<p className="truncate font-medium text-sm">
														{item.display_name}
													</p>
													<p className="text-muted-foreground text-xs">
														ilvl {item.item_level ?? "?"} ·{" "}
														{item.source === "bag" ? "Bag" : "Equipped"}
													</p>
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
