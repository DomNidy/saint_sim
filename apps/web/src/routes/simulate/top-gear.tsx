import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/simulate/top-gear")({
	component: RouteComponent,
});

function RouteComponent() {
	return <div>Hello "/simulate/top-gear"!</div>;
}
