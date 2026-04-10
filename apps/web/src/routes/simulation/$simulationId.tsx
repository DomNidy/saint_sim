import { createFileRoute } from "@tanstack/react-router";
import { getSimulationResult } from "@/lib/simulation.functions";

export const Route = createFileRoute("/simulation/$simulationId")({
	component: RouteComponent,
	loader: (loaderCtx) =>
		getSimulationResult({
			data: { simulationId: loaderCtx.params.simulationId },
		}),
});

function RouteComponent() {
	const { status, result } = Route.useLoaderData();
	return (
		<div>
			<p>Simulation status: {status}</p>
			<code>{JSON.stringify(result)}: </code>
		</div>
	);
}
