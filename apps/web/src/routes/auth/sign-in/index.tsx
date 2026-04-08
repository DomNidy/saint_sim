import { createFileRoute } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { authBrowserClient } from "@/lib/auth/auth-browser-client";

export const Route = createFileRoute("/auth/sign-in/")({
	component: RouteComponent,
});

function RouteComponent() {
	const signInWithDiscord = async () => {
		const data = await authBrowserClient.signIn.social({ provider: "discord" });
		console.log(`res data: ${JSON.stringify(data)}`);
	};
	return (
		<div>
			<Button onClick={signInWithDiscord}>Sign in w/ Discord</Button>
		</div>
	);
}
