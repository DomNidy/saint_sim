# Web app overview

This is the web front-end for interacting with the saint API. It is TypeScript application built with the [TanStack Start](https://tanstack.com/start) framework.

## shadcn instructions

Use the latest version of Shadcn to install new components, like this command to add a button component:

```bash
pnpm dlx shadcn@latest add button
```

## Authentication / Better Auth

This app uses Better Auth for authentication. We connect to the same postgres database as the other services.

The required datbase tables for Better Auth are applied manually. Better Auth provides a command to generate
a SQL migration script to create the necessary tables. This command is:

```
npx auth@latest generate
```

The resulting SQL migration file then needs to be added to `/db/migrations` in project root so it can be
applied on startup.

## Gotchas/Potential Issues 

Refer to TROUBLESHOOTING.md for common issues (and their solutions) that may occur throughout dev on the web app.

<!-- intent-skills:start -->
# Skill mappings - when working in these areas, load the linked skill file into context.
skills:
  - task: "changing the TanStack Start app setup, Vite plugin, router bootstrap, or root document shell"
    load: "node_modules/@tanstack/start-client-core/skills/start-core/SKILL.md"
  - task: "building or updating server functions in src/lib/*.functions.ts, including input validation and middleware"
    load: "node_modules/@tanstack/start-client-core/skills/start-core/server-functions/SKILL.md"
  - task: "working on protected routes, auth redirects, or Better Auth route guards in src/routes"
    load: "node_modules/@tanstack/router-core/skills/router-core/auth-and-guards/SKILL.md"
  - task: "changing server API handlers under src/routes/api"
    load: "node_modules/@tanstack/start-client-core/skills/start-core/server-routes/SKILL.md"
  - task: "editing route loaders, preload behavior, or route data fetching in src/routes and src/router.tsx"
    load: "node_modules/@tanstack/router-core/skills/router-core/data-loading/SKILL.md"
<!-- intent-skills:end -->
