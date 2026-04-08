import { createFileRoute } from "@tanstack/react-router";
import { authServerClient } from "@/lib/auth/auth-server-client";

export const Route = createFileRoute("/api/auth/$")({
	server: {
		handlers: {
			GET: ({ request }) => authServerClient.handler(request),
			POST: ({ request }) => authServerClient.handler(request),
		},
	},
});
