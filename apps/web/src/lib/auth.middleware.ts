import { createMiddleware } from "@tanstack/react-start";
import { getRequest } from "@tanstack/react-start/server";

import { auth } from "#/lib/auth";

export const requireAuthMiddleware = createMiddleware({
	type: "function",
}).server(async ({ next }) => {
	const session = await auth.api.getSession({
		headers: getRequest().headers,
	});

	if (!session?.user) {
		throw new Error("Authentication required.");
	}

	return next({
		context: {
			session,
			user: session.user,
		},
	});
});
