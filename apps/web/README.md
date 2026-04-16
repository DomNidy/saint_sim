# Web App

This app is the TanStack Start front-end for Saint. It handles Better Auth, web routes, and the browser client for the API server.

## Local dev

The repo uses `npm` for the web app.

From the repo root, the normal full-stack flow is:

```bash
just dev
just web
```

- `just dev` starts the Go services in Docker dev mode.
- `just web` runs this app locally on `http://localhost:3000`.

If you only need the web app plus auth/database access, the minimum dependency is Postgres plus this app:

```bash
docker compose up postgres
cd apps/web
npm install
npm run dev
```

## Environment

This app reads server env vars from `apps/web/.env`, validated in `src/env.ts`.

The important variables are:

- `BETTER_AUTH_URL`: canonical public origin for the web app. Keep this aligned with how you browse the app. `localhost` and `127.0.0.1` are different origins.
- `BETTER_AUTH_SECRET`: Better Auth signing secret.
- `DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_NAME`: Postgres connection for Better Auth.
- `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET`: Discord OAuth config.
- `SAINT_API_URL`, `SAINT_API_KEY`: API server base URL and auth used by the generated API client.

Local auth runs on the same app at `/api/auth`. See `src/lib/auth/auth.ts` for the current Better Auth wiring and trusted-origin behavior.

## Verify

Use the narrowest command that matches the change:

```bash
cd apps/web
npm run check -- src/path/to/file.tsx
npm run test -- src/path/to/file.test.tsx
npm run build
```

- `npm run check -- <paths>` is the default verification step for touched files.
- Add `npm run test -- <paths>` when interaction, state, or a bug fix deserves a focused test.
- Run `npm run build` for route, auth, env, loader/query, or server-side changes.

## Conventions

### Better Auth

- The app uses Better Auth against the same Postgres database as the rest of the stack.
- Better Auth migrations should end up in the repo-level Goose history under `db/migrations`, not left as app-local migration files.
- If auth fails with `INVALID_ORIGIN`, check `BETTER_AUTH_URL` first. Exact origin matching matters.

### TanStack Router + Query

- Prefer loader data for first render on SSR-prefetched routes.
- If a route also uses React Query after hydration, pass loader data as `initialData` so the client and server render the same initial tree.
- Avoid reading browser-only APIs like `localStorage` during render on routes that may SSR.

### Testing

- The app uses Vitest, Testing Library, and `jsdom`.
- `@testing-library/jest-dom` is not installed here, so use standard Vitest assertions unless you add that setup explicitly.
- Radix menus and dropdowns often open on `pointerDown`, not just `click`, so tests should match that behavior.

### Shadcn

This repo is `npm`-first. Do not assume `pnpm` is available.

To add a component, use an npm-friendly command such as:

```bash
cd apps/web
npx shadcn@latest add button
```

`components.json` in this directory is the local shadcn config.

## API client generation

The generated client lives under `src/lib/saint-api/generated`.

- Prefer `just codegen api` from the repo root when the API contract changes. That updates both the Go and web generated artifacts together.
- `cd apps/web && npm run codegen:api` only refreshes the web client.

## Troubleshooting

For common auth issues, see `TROUBLESHOOTING.md`.
