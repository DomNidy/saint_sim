import { createFileRoute, Outlet } from "@tanstack/react-router";

declare global {
	interface Window {
		// Added by the Wowhead tooltip script loaded in the route head.
		$WowheadPower?: {
			refreshLinks?: () => void;
		};
	}
}

// Configure the global Wowhead tooltip script before it loads:
// - `colorLinks: true` colors item links by quality.
// - `iconizeLinks: true` lets Wowhead prepend item icons to eligible links.
// - `renameLinks: false` keeps Wowhead from rewriting the link text.
const WOWHEAD_CONFIG_SCRIPT =
	"window.whTooltips={colorLinks:true,iconizeLinks:true,renameLinks:false};";

export const Route = createFileRoute("/simulate/")({
	component: () => <Outlet />,
	head: () => ({
		scripts: [
			{
				children: WOWHEAD_CONFIG_SCRIPT,
			},
			{
				src: "https://wow.zamimg.com/js/tooltips.js",
			},
		],
	}),
});
