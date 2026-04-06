import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { LoaderCircle, Search, Sparkles } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { useForm } from "react-hook-form";

import { Button } from "#/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "#/components/ui/card";
import {
	Form,
	FormControl,
	FormDescription,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "#/components/ui/form";
import { Input } from "#/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#/components/ui/select";
import {
	type WowCharacter as SimulationRequestInput,
	simulationRealms,
	simulationRegions,
	simulationRequestSchema,
} from "#/lib/saint-api/contracts";
import {
	getSimulationResultByRequestId,
	submitSimulationRequest,
} from "#/lib/simulation.functions";

export const Route = createFileRoute("/dashboard")({
	component: DashboardPage,
});

const POLL_INTERVAL_MS = 3000;
const POLL_TIMEOUT_MS = 2 * 60 * 1000;

type SimulationRequestStatus = "pending" | "complete" | "error" | "timed_out";

interface SimulationRequestHistoryItem {
	requestId: string;
	submittedAt: number;
	inputs: SimulationRequestInput;
	status: SimulationRequestStatus;
	result?: string;
	error?: string;
}

function DashboardPage() {
	const [requests, setRequests] = useState<SimulationRequestHistoryItem[]>([]);
	const [selectedRequestId, setSelectedRequestId] = useState<string | null>(
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
			requests.find((request) => request.requestId === selectedRequestId) ??
			null,
		[requests, selectedRequestId],
	);

	const submitMutation = useMutation({
		mutationFn: submitSimulationRequest,
		onMutate: () => {
			setSubmitError(null);
		},
		onSuccess: ({ simulationRequestId }, variables) => {
			const nextRequest: SimulationRequestHistoryItem = {
				requestId: simulationRequestId,
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
		if (!selectedRequestId) {
			return;
		}

		const request = requests.find(
			(entry) => entry.requestId === selectedRequestId,
		);
		if (!request || request.status !== "pending") {
			return;
		}

		const remainingMs = request.submittedAt + POLL_TIMEOUT_MS - Date.now();

		if (remainingMs <= 0) {
			setRequests((current) =>
				current.map((entry) =>
					entry.requestId === selectedRequestId
						? { ...entry, status: "timed_out" }
						: entry,
				),
			);
			return;
		}

		const timeoutId = window.setTimeout(() => {
			setRequests((current) =>
				current.map((entry) =>
					entry.requestId === selectedRequestId
						? { ...entry, status: "timed_out" }
						: entry,
				),
			);
		}, remainingMs);

		return () => window.clearTimeout(timeoutId);
	}, [requests, selectedRequestId]);

	const resultQuery = useQuery({
		queryKey: ["simulation-result", selectedRequestId],
		queryFn: () => {
			if (!selectedRequestId) {
				throw new Error(
					"A simulation request must be selected before polling.",
				);
			}

			return getSimulationResultByRequestId({
				data: { requestId: selectedRequestId },
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
		if (!selectedRequestId || !resultQuery.data) {
			return;
		}

		setRequests((current) =>
			current.map((entry) => {
				if (entry.requestId !== selectedRequestId) {
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
	}, [resultQuery.data, selectedRequestId]);

	useEffect(() => {
		if (!selectedRequestId || !resultQuery.isError) {
			return;
		}

		const message =
			resultQuery.error instanceof Error
				? resultQuery.error.message
				: "Unable to retrieve simulation status.";

		setRequests((current) =>
			current.map((entry) =>
				entry.requestId === selectedRequestId
					? { ...entry, status: "error", error: message }
					: entry,
			),
		);
	}, [resultQuery.error, resultQuery.isError, selectedRequestId]);

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

	return (
		<main className="page-wrap px-4 pb-10 pt-12">
			<section className="rise-in grid gap-6 lg:grid-cols-[minmax(0,1.15fr)_minmax(0,0.85fr)]">
				<Card className="relative overflow-hidden">
					<div className="pointer-events-none absolute -right-16 -top-16 h-40 w-40 rounded-full bg-[radial-gradient(circle,rgba(79,184,178,0.28),transparent_68%)]" />
					<CardHeader>
						<p className="island-kicker">Dashboard</p>
						<CardTitle>Submit a Saint simulation</CardTitle>
						<CardDescription>
							Submit a simulation request for your WoW Character.
						</CardDescription>
					</CardHeader>
					<CardContent className="space-y-6">
						<Form {...form}>
							<form
								className="space-y-5"
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
												<LoaderCircle className="animate-spin" />
												Sending request
											</>
										) : (
											<>
												<Sparkles />
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

						<div className="space-y-3">
							<div className="flex items-center justify-between gap-3">
								<div>
									<p className="island-kicker">Requests</p>
									<h2 className="text-lg font-semibold">In-memory history</h2>
								</div>
								<p className="text-xs text-(--sea-ink-soft)">
									Resets on refresh
								</p>
							</div>

							{requests.length ? (
								<div className="space-y-3">
									{requests.map((request) => (
										<button
											key={request.requestId}
											type="button"
											className={historyCardClassName(
												request.requestId === selectedRequestId,
											)}
											onClick={() => {
												setSelectedRequestId(request.requestId);
												setSubmitError(null);
											}}
										>
											<div className="flex items-start justify-between gap-3">
												<div className="min-w-0 space-y-1 text-left">
													<p className="font-semibold text-(--sea-ink)">
														{request.inputs.character_name}
													</p>
													<p className="text-sm text-(--sea-ink-soft)">
														{request.inputs.region.toUpperCase()} /{" "}
														{formatRealmLabel(request.inputs.realm)}
													</p>
												</div>
												<span
													className={historyStatusClassName(request.status)}
												>
													{formatHistoryStatus(request.status)}
												</span>
											</div>
											<p className="mt-3 truncate text-left text-xs text-(--sea-ink-soft)">
												{request.requestId}
											</p>
										</button>
									))}
								</div>
							) : (
								<div className="rounded-[1.5rem] border border-dashed border-(--line) bg-white/35 p-4 text-sm text-(--sea-ink-soft) dark:bg-[rgba(13,28,32,0.35)]">
									Your submitted simulation requests will appear here for this
									browser session.
								</div>
							)}
						</div>
					</CardContent>
				</Card>

				<Card className="rise-in" style={{ animationDelay: "90ms" }}>
					<CardHeader>
						<p className="island-kicker">Status</p>
						<CardTitle className="text-2xl">Simulation lifecycle</CardTitle>
						<CardDescription>
							The TanStack app receives the form submission, forwards it to the
							saint API, and polls for the completed result using request-id
							lookup.
						</CardDescription>
					</CardHeader>
					<CardContent className="space-y-5">
						<div className={statusPanelClassName(resultBody.tone)}>
							<div className="flex items-start gap-3">
								<div className="mt-0.5 rounded-full border border-current/20 bg-current/10 p-2">
									{isPolling ? (
										<LoaderCircle className="size-4 animate-spin" />
									) : (
										<Search className="size-4" />
									)}
								</div>
								<div className="min-w-0 flex-1 space-y-2">
									<h2 className="text-lg font-semibold">{resultBody.title}</h2>
									<p className="m-0 text-sm leading-6">
										{resultBody.description}
									</p>
								</div>
							</div>
						</div>

						<dl className="grid gap-3 rounded-3xl border border-(--line) bg-white/45 p-4 text-sm dark:bg-[rgba(13,28,32,0.55)]">
							<div className="flex items-center justify-between gap-3">
								<dt className="font-semibold text-(--sea-ink-soft)">
									Request id
								</dt>
								<dd className="text-right text-(--sea-ink)">
									{selectedRequest?.requestId ?? "Not selected yet"}
								</dd>
							</div>
							<div className="flex items-center justify-between gap-3">
								<dt className="font-semibold text-(--sea-ink-soft)">
									Polling status
								</dt>
								<dd className="text-right text-(--sea-ink)">
									{selectedRequest
										? formatHistoryStatus(selectedRequest.status)
										: "Idle"}
								</dd>
							</div>
						</dl>

						{selectedRequest?.status === "complete" ? (
							<div className="space-y-3">
								<p className="island-kicker">Simulation Report</p>
								<pre className="overflow-x-auto rounded-3xl border border-(--line) bg-[#1d2e45] p-4 text-sm leading-6 text-[#e8efff]">
									<code>{selectedRequest.result}</code>
								</pre>
							</div>
						) : null}
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

function historyCardClassName(isSelected: boolean) {
	return [
		"w-full rounded-[1.5rem] border p-4 transition text-left",
		isSelected
			? "border-[rgba(79,184,178,0.5)] bg-[rgba(79,184,178,0.12)] shadow-[0_18px_36px_rgba(17,42,58,0.08)]"
			: "border-(--line) bg-white/45 hover:border-[rgba(79,184,178,0.35)] hover:bg-white/70 dark:bg-[rgba(13,28,32,0.45)]",
	].join(" ");
}

function historyStatusClassName(status: SimulationRequestStatus) {
	switch (status) {
		case "complete":
			return "rounded-full bg-[rgba(47,106,74,0.12)] px-3 py-1 text-xs font-semibold text-[var(--palm)]";
		case "error":
			return "rounded-full bg-red-500/10 px-3 py-1 text-xs font-semibold text-red-700 dark:text-red-300";
		case "timed_out":
			return "rounded-full bg-amber-500/12 px-3 py-1 text-xs font-semibold text-amber-700 dark:text-amber-300";
		default:
			return "rounded-full bg-[rgba(79,184,178,0.12)] px-3 py-1 text-xs font-semibold text-[var(--lagoon-deep)]";
	}
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
		case "warning":
			return "rounded-[1.5rem] border border-amber-500/25 bg-amber-500/12 p-4 text-amber-700 dark:text-amber-300";
		default:
			return "rounded-[1.5rem] border border-[var(--line)] bg-white/55 p-4 text-[var(--sea-ink)] dark:bg-[rgba(13,28,32,0.55)]";
	}
}
