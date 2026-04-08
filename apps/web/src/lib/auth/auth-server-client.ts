import { betterAuth } from "better-auth";
import { tanstackStartCookies } from "better-auth/tanstack-start";
import { Pool } from "pg";
import { env } from "@/env";

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
export const authServerClient = betterAuth({
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
	plugins: [tanstackStartCookies()],
});
