import { createFileRoute, Link } from "@tanstack/react-router";

export const Route = createFileRoute("/simulate/")({
	component: RouteComponent,
});

function RouteComponent() {
	return (
		<div className="flex flex-col">
			<Link to="/simulate/basic">Basic simulation</Link>
			<Link to="/simulate/top-gear">Top gear simulation</Link>
		</div>
	);
}
