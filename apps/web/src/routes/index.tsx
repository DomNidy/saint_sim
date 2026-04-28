import { createFileRoute, Link } from "@tanstack/react-router";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";

export const Route = createFileRoute("/")({
	component: App,
});

function App() {
	return (
		<section className="w-full pb-8 pt-14">
			<Card>
				<CardHeader>
					<CardTitle>Basic</CardTitle>
					<CardDescription>Simulate your WoW Character</CardDescription>
				</CardHeader>
				<CardContent>
					<Link
						to="/simulate/basic"
						className="text-sm font-medium  underline underline-offset-4 transition-opacity hover:opacity-80"
					>
						Perform a basic simulation
					</Link>
				</CardContent>
			</Card>
			<Card>
				<CardHeader>
					<CardTitle>Top Gear</CardTitle>
					<CardDescription>Simulate your WoW Character</CardDescription>
				</CardHeader>
				<CardContent>
					<Link
						to="/simulate/top-gear"
						className="text-sm font-medium  underline underline-offset-4 transition-opacity hover:opacity-80"
					>
						Perform a top gear simulation
					</Link>
				</CardContent>
			</Card>
		</section>
	);
}
