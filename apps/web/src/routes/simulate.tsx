import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { LoaderCircle, Sparkles } from "lucide-react";
import { useForm } from "react-hook-form";

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
import { Input } from "@/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import {
	type WowCharacter as SimulationRequestInput,
	simulationRealms,
	simulationRegions,
	simulationRequestSchema,
} from "@/lib/saint-api/contracts";
import { submitSimulationRequest } from "@/lib/simulation.functions";

export const Route = createFileRoute("/simulate")({
	component: SimulationForm,
});

function SimulationForm() {
	const form = useForm<SimulationRequestInput>({
		resolver: zodResolver(simulationRequestSchema),
		defaultValues: {
			region: "us",
			realm: "hydraxis",
			character_name: "",
		},
	});

	const submitMutation = useMutation({
		mutationFn: submitSimulationRequest,
		// TODO: maybe redirect to sim page here /simulation/$id page
		// onSuccess: ({ simulationRequestId }, variables) => {},
	});

	return (
		<main className="page-wrap px-4 pb-10 pt-12">
			<Card className="relative overflow-hidden">
				<CardHeader className="gap-2">
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
									}}
								>
									Reset form
								</Button>
							</div>
						</form>
					</Form>
				</CardContent>
			</Card>
		</main>
	);
}

function formatRealmLabel(realm: string) {
	return realm.replace(/(^\w|-\w)/g, (match) =>
		match.replace("-", "").toUpperCase(),
	);
}
