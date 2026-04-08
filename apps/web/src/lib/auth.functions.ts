import { createServerFn } from "@tanstack/react-start";
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