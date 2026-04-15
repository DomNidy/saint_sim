import { createFileRoute, Link } from "@tanstack/react-router";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";

export const Route = createFileRoute("/")({ component: App });

function App() {
	return (
		<section className="w-full pb-8 pt-14">
			<Card>
				<CardHeader>
					<CardTitle>Simulate</CardTitle>
					<CardDescription>Simulate your WoW Character</CardDescription>
				</CardHeader>
				<CardContent>
					<Link
						to="/simulate"
						className="text-sm font-medium  underline underline-offset-4 transition-opacity hover:opacity-80"
					>
						Go to simulation
					</Link>
				</CardContent>
			</Card>
		</section>
	);
}
