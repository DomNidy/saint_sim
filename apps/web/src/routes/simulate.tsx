import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import {
	Clock3,
	LoaderCircle,
	Search,
	ShieldAlert,
	Sparkles,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { useForm } from "react-hook-form";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
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
	Empty,
	EmptyContent,
	EmptyDescription,
	EmptyHeader,
	EmptyMedia,
	EmptyTitle,
} from "@/components/ui/empty";
import {
	Form,
	FormControl,
	FormDescription,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Progress } from "@/components/ui/progress";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "@/components/ui/tooltip";
import {
	type WowCharacter as SimulationRequestInput,
	simulationRealms,
	simulationRegions,
	simulationRequestSchema,
} from "@/lib/saint-api/contracts";
import {
	getSimulationResult,
	submitSimulationRequest,
} from "@/lib/simulation.functions";

export const Route = createFileRoute("/simulate")({
	component: DashboardPage,
});

const POLL_INTERVAL_MS = 3000;
const POLL_TIMEOUT_MS = 2 * 60 * 1000;

type SimulationRequestStatus = "pending" | "complete" | "error" | "timed_out";

interface SimulationRequestHistoryItem {
	simulationId: string;
	submittedAt: number;
	inputs: SimulationRequestInput;
	status: SimulationRequestStatus;
	result?: string;
	error?: string;
}

function DashboardPage() {
	const [requests, setRequests] = useState<SimulationRequestHistoryItem[]>([]);
	const [selectedSimulationId, setSelectedRequestId] = useState<string | null>(
		null,
	);
	const [submitError, setSubmitError] = useState<string | null>(null);

	const form = useForm<SimulationRequestInput>({
		resolver: zodResolver(simulationRequestSchema),
		defaultValues: {
			region: "us",
			realm: "hydraxis",
			character_name: "",
		},
	});

	const selectedRequest = useMemo(
		() =>
			requests.find(
				(request) => request.simulationId === selectedSimulationId,
			) ?? null,
		[requests, selectedSimulationId],
	);

	const submitMutation = useMutation({
		mutationFn: submitSimulationRequest,
		onMutate: () => {
			setSubmitError(null);
		},
		onSuccess: ({ simulationRequestId }, variables) => {
			const nextRequest: SimulationRequestHistoryItem = {
				simulationId: simulationRequestId,
				submittedAt: Date.now(),
				inputs: variables.data,
				status: "pending",
			};

			setRequests((current) => [nextRequest, ...current]);
			setSelectedRequestId(simulationRequestId);
		},
		onError: (error) => {
			setSubmitError(
				error instanceof Error
					? error.message
					: "Unable to submit simulation request.",
			);
		},
	});

	useEffect(() => {
		if (!selectedSimulationId) {
			return;
		}

		const request = requests.find(
			(entry) => entry.simulationId === selectedSimulationId,
		);
		if (!request || request.status !== "pending") {
			return;
		}

		const remainingMs = request.submittedAt + POLL_TIMEOUT_MS - Date.now();

		if (remainingMs <= 0) {
			setRequests((current) =>
				current.map((entry) =>
					entry.simulationId === selectedSimulationId
						? { ...entry, status: "timed_out" }
						: entry,
				),
			);
			return;
		}

		const timeoutId = window.setTimeout(() => {
			setRequests((current) =>
				current.map((entry) =>
					entry.simulationId === selectedSimulationId
						? { ...entry, status: "timed_out" }
						: entry,
				),
			);
		}, remainingMs);

		return () => window.clearTimeout(timeoutId);
	}, [requests, selectedSimulationId]);

	const resultQuery = useQuery({
		queryKey: ["simulation-result", selectedSimulationId],
		queryFn: () => {
			if (!selectedSimulationId) {
				throw new Error(
					"A simulation request must be selected before polling.",
				);
			}

			return getSimulationResult({
				data: { simulationId: selectedSimulationId },
			});
		},
		enabled: selectedRequest?.status === "pending",
		refetchInterval: (query) => {
			if (selectedRequest?.status !== "pending") {
				return false;
			}

			if (
				query.state.data?.status === "complete" ||
				query.state.data?.status === "error"
			) {
				return false;
			}

			return POLL_INTERVAL_MS;
		},
		retry: false,
	});

	useEffect(() => {
		if (!selectedSimulationId || !resultQuery.data) {
			return;
		}

		setRequests((current) =>
			current.map((entry) => {
				if (entry.simulationId !== selectedSimulationId) {
					return entry;
				}

				if (resultQuery.data.status === "complete") {
					return {
						...entry,
						status: "complete",
						result:
							resultQuery.data.result.sim_result ??
							"The simulation completed, but the API did not return a report body.",
						error: undefined,
					};
				}

				if (resultQuery.data.status === "error") {
					return {
						...entry,
						status: "error",
						error:
							resultQuery.data.result.error_text ??
							"The simulation failed before a report was generated.",
					};
				}

				if (entry.status !== "timed_out") {
					return {
						...entry,
						status: "pending",
						error: undefined,
					};
				}

				return entry;
			}),
		);
	}, [resultQuery.data, selectedSimulationId]);

	useEffect(() => {
		if (!selectedSimulationId || !resultQuery.isError) {
			return;
		}

		const message =
			resultQuery.error instanceof Error
				? resultQuery.error.message
				: "Unable to retrieve simulation status.";

		setRequests((current) =>
			current.map((entry) =>
				entry.simulationId === selectedSimulationId
					? { ...entry, status: "error", error: message }
					: entry,
			),
		);
	}, [resultQuery.error, resultQuery.isError, selectedSimulationId]);

	const resultBody = useMemo(() => {
		if (submitError) {
			return {
				tone: "error" as const,
				title: "We could not start that simulation.",
				description: submitError,
			};
		}

		if (!selectedRequest) {
			return {
				tone: "neutral" as const,
				title: "Ready for a simulation request",
				description:
					"Pick a region and realm, enter a character name, and we will proxy the request through the web app to the Gin API.",
			};
		}

		if (selectedRequest.status === "timed_out") {
			return {
				tone: "warning" as const,
				title: "Simulation still running",
				description:
					"The request is still pending after two minutes. You can leave it in the list and check back later in this session.",
			};
		}

		if (selectedRequest.status === "error") {
			return {
				tone: "error" as const,
				title: "We hit an error while checking the report.",
				description:
					selectedRequest.error ?? "Unable to retrieve simulation status.",
			};
		}

		if (selectedRequest.status === "complete") {
			return {
				tone: "success" as const,
				title: "Simulation complete",
				description:
					selectedRequest.result ??
					"The simulation completed, but the API did not return a report body.",
			};
		}

		return {
			tone: "pending" as const,
			title: "Simulation queued",
			description:
				"We forwarded your request through TanStack Start and are polling the saint API for the completed report.",
		};
	}, [selectedRequest, submitError]);

	const isPolling =
		selectedRequest?.status === "pending" &&
		!resultQuery.isError &&
		resultQuery.data?.status !== "complete";

	const progressValue = selectedRequest
		? selectedRequest.status === "pending"
			? Math.min(
					95,
					Math.max(
						8,
						Math.round(
							((Date.now() - selectedRequest.submittedAt) / POLL_TIMEOUT_MS) *
								100,
						),
					),
				)
			: selectedRequest.status === "timed_out"
				? 100
				: 100
		: 0;

	return (
		<main className="page-wrap px-4 pb-10 pt-12">
			<section className="rise-in grid gap-6 lg:grid-cols-[minmax(0,1.15fr)_minmax(0,0.85fr)]">
				<Card className="relative overflow-hidden">
					<div className="pointer-events-none absolute -right-16 -top-16 h-40 w-40 rounded-full bg-[radial-gradient(circle,rgba(79,184,178,0.28),transparent_68%)]" />
					<CardHeader className="gap-2">
						<p className="island-kicker">Dashboard</p>
						<CardTitle>Submit a Saint simulation</CardTitle>
						<CardDescription>
							Submit a simulation request for your WoW Character.
						</CardDescription>
					</CardHeader>
					<CardContent className="flex flex-col gap-6">
						<Form {...form}>
							<form
								className="flex flex-col gap-5"
								onSubmit={form.handleSubmit((values) => {
									void submitMutation.mutateAsync({ data: values });
								})}
							>
								<div className="grid gap-5 md:grid-cols-2">
									<FormField
										control={form.control}
										name="region"
										render={({ field }) => (
											<FormItem>
												<FormLabel>Region</FormLabel>
												<FormControl>
													<Select
														onValueChange={field.onChange}
														value={field.value}
													>
														<SelectTrigger>
															<SelectValue placeholder="Select a region" />
														</SelectTrigger>
														<SelectContent>
															{simulationRegions.map((region) => (
																<SelectItem key={region} value={region}>
																	{region.toUpperCase()}
																</SelectItem>
															))}
														</SelectContent>
													</Select>
												</FormControl>
												<FormDescription>
													The Blizzard region for the character&apos;s realm.
												</FormDescription>
												<FormMessage />
											</FormItem>
										)}
									/>

									<FormField
										control={form.control}
										name="realm"
										render={({ field }) => (
											<FormItem>
												<FormLabel>Realm</FormLabel>
												<FormControl>
													<Select
														onValueChange={field.onChange}
														value={field.value}
													>
														<SelectTrigger>
															<SelectValue placeholder="Select a realm" />
														</SelectTrigger>
														<SelectContent>
															{simulationRealms.map((realm) => (
																<SelectItem key={realm} value={realm}>
																	{formatRealmLabel(realm)}
																</SelectItem>
															))}
														</SelectContent>
													</Select>
												</FormControl>
												<FormDescription>
													Realms are limited to the current API contract.
												</FormDescription>
												<FormMessage />
											</FormItem>
										)}
									/>
								</div>

								<FormField
									control={form.control}
									name="character_name"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Character name</FormLabel>
											<FormControl>
												<Input
													placeholder="Thrallslayer"
													autoComplete="off"
													{...field}
												/>
											</FormControl>
											<FormDescription>
												Use the exact in-game character name for the selected
												realm.
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
											form.reset();
											setSubmitError(null);
										}}
									>
										Reset form
									</Button>
								</div>
							</form>
						</Form>

						<Separator />

						<div className="flex flex-col gap-3">
							<div className="flex items-center justify-between gap-3">
								<div className="flex flex-col gap-1">
									<p className="island-kicker">Requests</p>
									<h2 className="text-lg font-semibold">In-memory history</h2>
								</div>
								<p className="text-xs text-(--sea-ink-soft)">
									Resets on refresh
								</p>
							</div>

							{requests.length ? (
								<ScrollArea className="h-112 rounded-3xl border border-(--line) bg-white/35 p-1 dark:bg-[rgba(13,28,32,0.35)]">
									<div className="flex flex-col gap-3 p-3">
										{requests.map((request) => (
											<button
												key={request.simulationId}
												type="button"
												className={historyCardClassName(
													request.simulationId === selectedSimulationId,
												)}
												onClick={() => {
													setSelectedRequestId(request.simulationId);
													setSubmitError(null);
												}}
											>
												<div className="flex items-start justify-between gap-3">
													<div className="min-w-0 flex-1 text-left">
														<div className="flex flex-wrap items-start gap-2">
															<Link
																to="/simulation/$simulationId"
																params={{
																	simulationId: request.simulationId,
																}}
																className={simulationLinkClassName()}
																aria-label={`Open simulation page for ${request.inputs.character_name}`}
																onClick={(event) => {
																	event.stopPropagation();
																}}
															>
																{request.inputs.character_name}
															</Link>
														</div>
														<p className="text-sm text-(--sea-ink-soft)">
															{request.inputs.region.toUpperCase()} /{" "}
															{formatRealmLabel(request.inputs.realm)}
														</p>
													</div>
													<Badge
														variant={historyStatusBadgeVariant(request.status)}
													>
														{formatHistoryStatus(request.status)}
													</Badge>
												</div>
												<Separator className="my-3" />
												<div className="flex items-center justify-between gap-3">
													<p className="text-xs text-(--sea-ink-soft)">
														Submitted {formatTimestamp(request.submittedAt)}
													</p>
													<Tooltip>
														<TooltipTrigger asChild>
															<span className="max-w-56 truncate text-right text-xs text-(--sea-ink-soft)">
																{formatRequestId(request.simulationId)}
															</span>
														</TooltipTrigger>
														<TooltipContent>
															{request.simulationId}
														</TooltipContent>
													</Tooltip>
												</div>
											</button>
										))}
									</div>
								</ScrollArea>
							) : (
								<Empty className="min-h-72 rounded-3xl border border-dashed border-(--line) bg-white/35 dark:bg-[rgba(13,28,32,0.35)]">
									<EmptyContent>
										<EmptyMedia variant="icon">
											<Sparkles />
										</EmptyMedia>
										<EmptyHeader>
											<EmptyTitle>No requests yet</EmptyTitle>
											<EmptyDescription>
												Your submitted simulation requests will appear here for
												this browser session.
											</EmptyDescription>
										</EmptyHeader>
									</EmptyContent>
								</Empty>
							)}
						</div>
					</CardContent>
				</Card>

				<Card className="rise-in" style={{ animationDelay: "90ms" }}>
					<CardHeader className="gap-4">
						<div className="flex flex-wrap items-start justify-between gap-4">
							<div className="flex flex-col gap-2">
								<p className="island-kicker">Status</p>
								<CardTitle className="text-2xl">Simulation lifecycle</CardTitle>
								<CardDescription>
									The TanStack app receives the form submission, forwards it to
									the saint API, and polls for the completed result using
									request-id lookup.
								</CardDescription>
							</div>
						</div>

						<Separator />
					</CardHeader>
					<CardContent className="flex flex-col gap-6">
						<Tabs defaultValue="summary" className="w-full">
							<TabsList variant="line" className="w-full justify-start">
								<TabsTrigger value="summary">Summary</TabsTrigger>
								<TabsTrigger value="report">Report</TabsTrigger>
							</TabsList>

							<TabsContent value="summary" className="pt-4">
								<div className="flex flex-col gap-4">
									{resultBody.tone === "error" ? (
										<Alert variant="destructive">
											<ShieldAlert />
											<AlertTitle>{resultBody.title}</AlertTitle>
											<AlertDescription>
												{resultBody.description}
											</AlertDescription>
										</Alert>
									) : resultBody.tone === "warning" ? (
										<Alert>
											<Clock3 />
											<AlertTitle>{resultBody.title}</AlertTitle>
											<AlertDescription>
												{resultBody.description}
											</AlertDescription>
										</Alert>
									) : (
										<div className={statusPanelClassName(resultBody.tone)}>
											<div className="flex items-start gap-3">
												<div className="mt-0.5 rounded-full border border-current/20 bg-current/10 p-2">
													{isPolling ? (
														<LoaderCircle
															data-icon="inline-start"
															className="animate-spin"
														/>
													) : (
														<Search data-icon="inline-start" />
													)}
												</div>
												<div className="min-w-0 flex-1">
													<div className="flex flex-wrap items-center gap-2">
														<h2 className="text-lg font-semibold">
															{resultBody.title}
														</h2>
														{selectedRequest ? (
															<Badge
																variant={historyStatusBadgeVariant(
																	selectedRequest.status,
																)}
															>
																{formatHistoryStatus(selectedRequest.status)}
															</Badge>
														) : null}
													</div>
													<p className="mt-2 text-sm leading-6">
														{resultBody.description}
													</p>
												</div>
											</div>

											<Separator className="my-4" />

											<div className="flex flex-col gap-2">
												<div className="flex items-center justify-between gap-3">
													<p className="text-sm font-medium">
														Polling progress
													</p>
													<p className="text-xs text-(--sea-ink-soft)">
														{selectedRequest?.status === "pending"
															? `${formatElapsedMs(
																	Date.now() - selectedRequest.submittedAt,
																)} elapsed`
															: selectedRequest?.status === "timed_out"
																? "Timed out"
																: selectedRequest?.status === "complete"
																	? "Complete"
																	: "Idle"}
													</p>
												</div>
												<Progress value={progressValue} />
											</div>
										</div>
									)}
								</div>
							</TabsContent>

							<TabsContent value="report" className="pt-4">
								{selectedRequest?.status === "complete" ? (
									<div className="flex flex-col gap-4">
										<div className="flex items-center justify-between gap-3">
											<div className="flex flex-col gap-1">
												<p className="text-sm font-medium">Raw report</p>
												<p className="text-xs text-(--sea-ink-soft)">
													The full result is scrollable below and can also be
													opened in a dialog.
												</p>
											</div>
											<Badge variant="secondary">Ready</Badge>
										</div>
										<ScrollArea className="h-104 rounded-3xl border border-(--line) bg-[#1d2e45] p-4">
											<pre className="m-0 whitespace-pre-wrap text-sm leading-6 text-[#e8efff]">
												<code>{selectedRequest.result}</code>
											</pre>
										</ScrollArea>
									</div>
								) : selectedRequest?.status === "pending" ? (
									<div className="flex flex-col gap-4 rounded-3xl border border-(--line) bg-white/55 p-4 dark:bg-[rgba(13,28,32,0.55)]">
										<div className="flex items-center gap-3">
											<LoaderCircle
												data-icon="inline-start"
												className="animate-spin"
											/>
											<p className="text-sm font-medium">
												The report is still being generated.
											</p>
										</div>
										<Skeleton className="h-4 w-full" />
										<Skeleton className="h-4 w-5/6" />
										<Skeleton className="h-4 w-2/3" />
										<Progress value={progressValue} />
									</div>
								) : selectedRequest?.status === "error" ? (
									<Alert variant="destructive">
										<ShieldAlert />
										<AlertTitle>Report unavailable</AlertTitle>
										<AlertDescription>
											{selectedRequest.error ??
												"Unable to retrieve simulation status."}
										</AlertDescription>
									</Alert>
								) : (
									<Empty className="min-h-64 rounded-3xl border border-dashed border-(--line) bg-white/35 dark:bg-[rgba(13,28,32,0.35)]">
										<EmptyContent>
											<EmptyMedia variant="icon">
												<Search />
											</EmptyMedia>
											<EmptyHeader>
												<EmptyTitle>No report yet</EmptyTitle>
												<EmptyDescription>
													Select a completed simulation to see the raw report
													body.
												</EmptyDescription>
											</EmptyHeader>
										</EmptyContent>
									</Empty>
								)}
							</TabsContent>
						</Tabs>
					</CardContent>
				</Card>
			</section>
		</main>
	);
}

function formatRealmLabel(realm: string) {
	return realm.replace(/(^\w|-\w)/g, (match) =>
		match.replace("-", "").toUpperCase(),
	);
}

function formatHistoryStatus(status: SimulationRequestStatus) {
	switch (status) {
		case "complete":
			return "Complete";
		case "error":
			return "Error";
		case "timed_out":
			return "Timed out";
		default:
			return "Checking for report";
	}
}

function formatTimestamp(value: number) {
	return new Intl.DateTimeFormat("en-US", {
		month: "short",
		day: "numeric",
		hour: "numeric",
		minute: "2-digit",
	}).format(value);
}

function formatElapsedMs(value: number) {
	if (value < 1000) {
		return "Less than 1s";
	}

	const totalSeconds = Math.floor(value / 1000);
	const minutes = Math.floor(totalSeconds / 60);
	const seconds = totalSeconds % 60;

	if (minutes > 0) {
		return `${minutes}m ${seconds}s`;
	}

	return `${seconds}s`;
}

function formatRequestId(requestId: string) {
	if (requestId.length <= 18) {
		return requestId;
	}

	return `${requestId.slice(0, 8)}…${requestId.slice(-6)}`;
}

function historyStatusBadgeVariant(status: SimulationRequestStatus) {
	switch (status) {
		case "complete":
			return "secondary" as const;
		case "error":
			return "destructive" as const;
		case "timed_out":
			return "outline" as const;
		default:
			return "default" as const;
	}
}

function historyCardClassName(isSelected: boolean) {
	return [
		"w-full rounded-[1.5rem] border p-4 text-left transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
		isSelected
			? "border-[rgba(79,184,178,0.5)] bg-[rgba(79,184,178,0.12)] shadow-[0_18px_36px_rgba(17,42,58,0.08)]"
			: "border-(--line) bg-white/45 hover:border-[rgba(79,184,178,0.35)] hover:bg-white/70 dark:bg-[rgba(13,28,32,0.45)]",
	].join(" ");
}

function statusPanelClassName(
	tone: "error" | "neutral" | "pending" | "success" | "warning",
) {
	switch (tone) {
		case "error":
			return "rounded-[1.5rem] border border-red-500/20 bg-red-500/10 p-4 text-red-700 dark:text-red-300";
		case "pending":
			return "rounded-[1.5rem] border border-[rgba(79,184,178,0.25)] bg-[rgba(79,184,178,0.12)] p-4 text-[var(--lagoon-deep)]";
		case "success":
			return "rounded-[1.5rem] border border-[rgba(47,106,74,0.22)] bg-[rgba(47,106,74,0.12)] p-4 text-[var(--palm)] dark:text-[var(--palm)]";
		default:
			return "rounded-[1.5rem] border border-[var(--line)] bg-white/55 p-4 text-[var(--sea-ink)] dark:bg-[rgba(13,28,32,0.55)]";
	}
}

function simulationLinkClassName() {
	return [
		"inline font-semibold text-(--sea-ink) underline underline-offset-4 transition-colors hover:text-[var(--lagoon-deep)] focus-visible:outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50",
	].join(" ");
}
