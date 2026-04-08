import { createServerFn, createServerOnlyFn } from "@tanstack/react-start";
import { auth } from "./auth";
import { getRequestHeaders } from "@tanstack/react-start/server";
import { createRemoteJWKSet, jwtVerify } from "jose"
import { env } from "@/env";

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


/**
 * Server function to verify that a JWT is valid for the saint API.
 * Call this before sending a JWT to the saint API to ensure we
 * don't send an invalid token to it.
 * 
 * We validate two claims:
 * - issuer: Token must be issued by web (BETTER_AUTH_URL)
 * - audience: Token audience should be saint api URL (token is meant
 * only to authenticate w/ saint api)
 */
export const verifyJwtForSaintApi = createServerOnlyFn(async (token: string) => {
    const JWKS = createRemoteJWKSet(new URL('http://localhost:3000/api/auth/jwks'))
    try {
        const { payload } = await jwtVerify(token, JWKS, {
            issuer: env.BETTER_AUTH_URL,
            audience: env.SAINT_API_URL,
        })
        return payload !== null
    } catch (exception) {
        console.log(`Failed to validate JWT: ${exception}`)
    }
    return false
})