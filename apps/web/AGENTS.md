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
