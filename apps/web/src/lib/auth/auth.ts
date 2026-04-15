import { betterAuth } from "better-auth";
import { jwt } from "better-auth/plugins";
import { tanstackStartCookies } from "better-auth/tanstack-start";
import { Pool } from "pg";
import { env } from "@/env";

function trustedOriginsFor(baseURL: string): string[] {
	const configuredOrigin = new URL(baseURL).origin;
	const trustedOrigins = new Set([configuredOrigin]);
	const url = new URL(configuredOrigin);

	// In local development, Better Auth sees localhost and 127.0.0.1 as
	// different origins even when they resolve to the same server.
	if (url.hostname === "localhost") {
		trustedOrigins.add(url.origin.replace("localhost", "127.0.0.1"));
	}

	if (url.hostname === "127.0.0.1") {
		trustedOrigins.add(url.origin.replace("127.0.0.1", "localhost"));
	}

	return [...trustedOrigins];
}

/**
 * Better auth server-side client.
 *
 * Not meant to be called from browser/client-side;
 * only call on server.
 *
 * NOTE: If you try to access this on client, it
 * browser console might log error due to node-only
 * types being pulled in (e.g., from pg)
 */
export const auth = betterAuth({
	baseURL: env.BETTER_AUTH_URL,
	trustedOrigins: trustedOriginsFor(env.BETTER_AUTH_URL),
	emailAndPassword: {
		enabled: true,
	},
	database: new Pool({
		user: env.DB_USER,
		password: env.DB_PASSWORD,
		host: env.DB_HOST,
		database: env.DB_NAME,
		port: 5432,
	}),
	socialProviders: {
		discord: {
			clientId: env.DISCORD_CLIENT_ID,
			clientSecret: env.DISCORD_CLIENT_SECRET,
			// permissions: ... - permissions param works only when `bot` scope is
			// included in the OAuth2 scopes. We might wanna look into this later,
			// cause we may find a use for it.
			// link: https://docs.discord.com/developers/topics/permissions
		},
	},
	plugins: [jwt(), tanstackStartCookies()],
});
