import { TanStackDevtools } from "@tanstack/react-devtools";
import type { QueryClient } from "@tanstack/react-query";
import { ReactQueryDevtoolsPanel } from "@tanstack/react-query-devtools";
import {
	createRootRouteWithContext,
	HeadContent,
	Scripts,
} from "@tanstack/react-router";
import { TanStackRouterDevtoolsPanel } from "@tanstack/react-router-devtools";
import type { Auth } from "better-auth/types";
import TanStackQueryProvider from "@/providers/tanstack-query-provider";
import Footer from "../components/footer";
import Header from "../components/header";
import { TooltipProvider } from "../components/ui/tooltip";
import appCss from "../styles.css?url";

/**
 * Context object to pass things along routes
 */
interface AppRouterContext {
	queryClient: QueryClient;
	session?: Awaited<ReturnType<Auth["api"]["getSession"]>>;
}

const THEME_INIT_SCRIPT = `(function(){try{var stored=window.localStorage.getItem('theme');var mode=(stored==='light'||stored==='dark'||stored==='auto')?stored:'auto';var prefersDark=window.matchMedia('(prefers-color-scheme: dark)').matches;var resolved=mode==='auto'?(prefersDark?'dark':'light'):mode;var root=document.documentElement;root.classList.remove('light','dark');root.classList.add(resolved);if(mode==='auto'){root.removeAttribute('data-theme')}else{root.setAttribute('data-theme',mode)}root.style.colorScheme=resolved;}catch(e){}})();`;

const WOWHEAD_CONFIG_SCRIPT =
	"window.whTooltips={colorLinks:true,iconizeLinks:true,renameLinks:false};";

const WOWHEAD_SCRIPT_SRC = "https://wow.zamimg.com/js/tooltips.js";

export const Route = createRootRouteWithContext<AppRouterContext>()({
	head: () => ({
		meta: [
			{
				charSet: "utf-8",
			},
			{
				name: "viewport",
				content: "width=device-width, initial-scale=1",
			},
			{
				title: "Saint Sim",
			},
		],
		links: [
			{
				rel: "stylesheet",
				href: appCss,
			},
		],
		scripts: [
			{
				children: WOWHEAD_CONFIG_SCRIPT,
			},
			{
				src: WOWHEAD_SCRIPT_SRC,
			},
		],
	}),
	shellComponent: RootDocument,
});

function RootDocument({ children }: { children: React.ReactNode }) {
	return (
		<html lang="en" suppressHydrationWarning>
			<head>
				{/** biome-ignore lint/security/noDangerouslySetInnerHtml: dark mode script */}
				<script dangerouslySetInnerHTML={{ __html: THEME_INIT_SCRIPT }} />
				<HeadContent />
			</head>
			<body className="font-sans antialiased wrap-anywhere">
				<TooltipProvider>
					<TanStackQueryProvider>
						<Header />
						<main className="mx-auto flex w-full max-w-6xl flex-col px-4 sm:px-6 lg:px-8">
							{children}
						</main>
						<Footer />
						<TanStackDevtools
							config={{
								position: "bottom-right",
							}}
							plugins={[
								{
									name: "Tanstack Router",
									render: <TanStackRouterDevtoolsPanel />,
								},
								{
									name: "Tanstack Query",
									render: <ReactQueryDevtoolsPanel />,
								},
							]}
						/>
					</TanStackQueryProvider>
				</TooltipProvider>
				<Scripts />
			</body>
		</html>
	);
}
