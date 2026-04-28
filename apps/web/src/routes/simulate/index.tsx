import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/simulate/")({
	beforeLoad: () => {
		throw redirect({
			to: "/simulate/basic",
			replace: true,
		});
	},
	component: () => <Outlet />,
});
