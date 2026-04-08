import { createMiddleware } from "@tanstack/react-start";
import { getRequest } from "@tanstack/react-start/server";

import { auth } from "@/lib/auth/auth-server-client";

/**
 * Middleware that ensures incoming request is authenticated.
 * Add this to server functions, or server routes that require
 * authentication.
 *
 * NOTE: This is a request middleware (default option when calling
 * `createMiddleware`)
 * 
 * Tanstack start offers two middleware types:
 * 
 * - Request middleware (customize the behavior of ANY
 * server request that passes through it, including
 * server routes, SSR, and server functions)
 *
 * - Server function middleware that run only on
 * server functions (you can configure which server
 * functions use a middleware w/ .middleware([mw1,...]))
 * 
 * DOCS: https://tanstack.com/start/latest/docs/framework/react/guide/middleware#request-middleware
 */
export const requireAuthMiddleware = createMiddleware().
	server(async ({ next }) => {
		// get headers of req so we can extract
		// better auth session token from 'Cookie' header
		const headers = getRequest().headers

		// better auth checks that the session token
		// is valid, looking up the token in database,
		// and resolving to a session w/ data if so
		const session = await auth.api.getSession({
			headers: headers
		});

		if (!session?.user) {
			throw new Error("Authentication required.");
		}

		// aggregate the session data in request context
		// so subsequent server fns has access to it
		return next({
			context: {
				session,
				user: session.user,
			},
		});
	});
