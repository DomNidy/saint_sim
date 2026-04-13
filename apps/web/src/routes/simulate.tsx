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
import { Textarea } from "@/components/ui/textarea";
import {
	simulationRequestSchema,
	type SimulationRequestInput,
} from "@/lib/saint-api/contracts";
import { submitSimulationRequest } from "@/lib/simulation.functions";

export const Route = createFileRoute("/simulate")({
	component: SimulationForm,
});

function SimulationForm() {
	const form = useForm<SimulationRequestInput>({
		resolver: zodResolver(simulationRequestSchema),
		defaultValues: {
			simc_addon_export: "",
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
						Paste your SimulationCraft addon export to submit a simulation.
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
							<FormField
								control={form.control}
								name="simc_addon_export"
								render={({ field }) => (
									<FormItem>
										<FormLabel>SimC addon export</FormLabel>
										<FormControl>
											<Textarea
												placeholder={
													'priest="Example"\nlevel=80\nspec=shadow'
												}
												autoComplete="off"
												rows={14}
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
