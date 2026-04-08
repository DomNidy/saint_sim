import { createServerFn, createServerOnlyFn } from "@tanstack/react-start";
import { auth } from "./auth";
import { getRequestHeaders } from "@tanstack/react-start/server";

/**
 * Server function to get the session from user.
 */
export const getSession = createServerFn({ method: "GET" }).handler(async () => {
    const headers = getRequestHeaders();
    const session = await auth.api.getSession({ headers })

    return session
})

/**
 * Server function to ensure the user has a session, throws error if not
 */
export const ensureSession = createServerFn({ method: "GET" }).handler(async () => {
    const headers = getRequestHeaders();
    const session = await auth.api.getSession({ headers })

    if (!session) {
        throw new Error("Unauthorized")
    }

    return session
})

/**
 * Server only function to retrieve better auth JWT
 * 
 * Server only because client side token retrieval is best done
 * using authClient.token()
 * 
 * Docs: https://better-auth.com/docs/plugins/jwt#retrieve-the-token
 */
export const getToken = createServerOnlyFn(async () => {
    const token = await auth.api.getToken({
        headers: getRequestHeaders()
    })
    return token
})