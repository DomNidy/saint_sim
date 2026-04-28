import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { TopGearSimulationResultDisplay } from "@/components/simulation-results/top-gear-result";
import type { GetSimulationResponse } from "@/lib/saint-api/generated";
import { isResult } from "@/lib/simulation/result";
import { getSimulationResult } from "@/lib/simulation.functions";

const simulationQueryOptions = (simulationId: string) =>
	queryOptions({
		queryKey: ["simulation", simulationId],
		queryFn: () =>
			getSimulationResult({ data: { simulationId: simulationId } }),
	});

export const Route = createFileRoute("/simulation/$simulationId")({
	component: RouteComponent,
	loader: ({ context, params }) =>
		context.queryClient.ensureQueryData(
			simulationQueryOptions(params.simulationId),
		),
});

const SimulationLogViewer = ({ sim }: { sim: GetSimulationResponse }) => {
	return (
		<div>
			{isResult(sim.result, "topGear") && (
				<div>
					<TopGearSimulationResultDisplay
						equipment={sim.result.equipment}
						kind="topGear"
						metric={sim.result.metric}
						profilesets={sim.result.profilesets}
					/>
				</div>
			)}
			<p>Status: {sim?.status ?? "unknown"}</p>

			{sim?.status === "error" && <code>{sim?.error_text}</code>}
			{sim?.status === "complete" && (
				<div>
					<h2 className="font-bold text-xl">Raw result</h2>
					<div className="bg-secondary h-96 overflow-y-scroll">
						<code>{JSON.stringify(sim?.result ?? "{}")}</code>
					</div>
				</div>
			)}

			{sim?.raw_simc_input && (
				<div>
					<h2 className="font-bold text-xl">Raw simc profile</h2>

					<div className="bg-secondary h-96 overflow-y-scroll">
						{sim?.raw_simc_input?.split("\n").map((v) => (
							<p key={v}>{v}</p>
						))}
					</div>
				</div>
			)}
		</div>
	);
};

function RouteComponent() {
	const { simulationId } = Route.useParams();
	const initialSimulation = Route.useLoaderData();
	const simulation = useQuery({
		...simulationQueryOptions(simulationId),
		initialData: initialSimulation,
		refetchInterval: (query) =>
			query.state.data?.status === "in_queue" ||
			query.state.data?.status === "in_progress"
				? 2_000
				: false,
	});

	return (
		<div>
			<SimulationLogViewer sim={simulation.data} />
		</div>
	);
}
