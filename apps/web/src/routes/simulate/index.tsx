import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";
import {
	WOWHEAD_CONFIG_SCRIPT,
	WOWHEAD_SCRIPT_SRC,
} from "@/lib/equipment/wowhead";

export const Route = createFileRoute("/simulate/")({
	beforeLoad: () => {
		throw redirect({
			to: "/simulate/basic",
			replace: true,
		});
	},
	component: () => <Outlet />,
	head: () => ({
		scripts: [
			{
				children: WOWHEAD_CONFIG_SCRIPT,
			},
			{
				src: WOWHEAD_SCRIPT_SRC,
			},
		],
	}),
});
