import { betterAuth } from "better-auth";
import { tanstackStartCookies } from "better-auth/tanstack-start";
import { Pool } from "pg";

import { env } from "#/env";

export const auth = betterAuth({
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
			clientSecret: env.DISCORD_CLIENT_SECRET
		}
	},
	plugins: [tanstackStartCookies()],
});
