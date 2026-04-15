import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
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

function RouteComponent() {
	const { simulationId } = Route.useParams();
	const initialSimulation = Route.useLoaderData();
	const simulation = useQuery({
		...simulationQueryOptions(simulationId),
		initialData: initialSimulation,
		refetchInterval: (query) =>
			query.state.data?.status === "pending" ? 2_000 : false,
	});

	return (
		<div>
			<p>Simulation status: {simulation.data.status}</p>
			<code>{JSON.stringify(simulation.data.result)}</code>
		</div>
	);
}
