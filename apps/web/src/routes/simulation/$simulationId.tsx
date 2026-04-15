import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import type { GetSimulationResponse } from "@/lib/saint-api/generated";
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
			<p>Status: {sim?.simulation_status ?? "unknown"}</p>
			<code>{sim?.sim_result}</code>
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
			query.state.data?.simulation_status === "in_queue" ||
			query.state.data?.simulation_status === "in_progress"
				? 2_000
				: false,
	});

	return (
		<div>
			<SimulationLogViewer sim={simulation.data} />
		</div>
	);
}
