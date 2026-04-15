import { Link } from "@tanstack/react-router";
import BetterAuthHeader from "./auth/header-user.tsx";
import ThemeToggle from "./theme-toggle.tsx";

export default function Header() {
	return (
		<header className="sticky top-0 z-50 border-b border-border/80 bg-background/80 px-3 backdrop-blur-lg sm:px-4">
			<nav className="page-wrap mx-auto flex w-full max-w-6xl flex-wrap items-center justify-between gap-x-3 gap-y-2 py-3 sm:py-4">
				<div className="flex flex-1 flex-wrap items-center gap-x-2 gap-y-1 pt-1 text-sm font-medium sm:gap-x-3">
					<Link
						className="mr-2 text-2xl font-bold tracking-tight text-foreground"
						to="/"
					>
						Saint
					</Link>

					<Link
						to="/"
						className="inline-flex h-9 items-center rounded-md px-3 text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
						activeProps={{
							className:
								"inline-flex h-9 items-center rounded-md bg-secondary px-3 text-secondary-foreground",
						}}
					>
						Home
					</Link>
					<Link
						to="/simulate"
						className="inline-flex h-9 items-center rounded-md px-3 text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
						activeProps={{
							className:
								"inline-flex h-9 items-center rounded-md bg-secondary px-3 text-secondary-foreground",
						}}
					>
						Simulate
					</Link>
				</div>
				<div className="flex items-center gap-2">
					<BetterAuthHeader />
					<ThemeToggle />
				</div>
			</nav>
		</header>
	);
}
