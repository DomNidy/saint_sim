# Repo Guidance

## Change -> Verify

| Change | Minimum verify |
| --- | --- |
| Web UI/component change | `cd apps/web && npm run check -- <paths>` |
| Web route, auth, loader/query, or env change | `cd apps/web && npm run check -- <paths>` and `cd apps/web && npm run build` |
| Web interaction/state change | `cd apps/web && npm run check -- <paths>` and `cd apps/web && npm run test -- <paths>` when a focused test exists or is added |
| Go handler/service/shared package change | Run narrow `golangci-lint` in the changed directory or file, then a targeted `GOCACHE=/tmp/go-build go test ...` or `go build ...` for the touched package |
| Docker/dev image/compose change | `docker compose -f docker-compose.yml config --services` and, for dev-stack changes, `docker compose -f docker-compose.yml -f docker-compose.dev.yml config --services`; build only the touched service if Docker is available |
| Codegen / OpenAPI contract change | Prefer `just codegen api` so Go and web generated types stay aligned; use `just codegen db` only when the DB/sqlc path is actually available |
| Migration / schema change | Start Postgres, run `just db-migrate`, then rerun the affected codegen/tests |

## Go linting

- Repo-wide `golangci-lint run ./...` works, but prefer per-directory or per-file runs for day-to-day changes so results stay scoped to the code you touched.
- Prefer narrow commands for only the paths you changed, for example:
  - `cd apps/api/handlers && golangci-lint run`
  - `cd apps/api && golangci-lint run main.go`
- Use repo-wide lint intentionally when you want whole-repo signal. It will include unrelated existing findings and is slower/noisier than a narrow run.
- Treat cache persistence warnings as non-blocking noise unless they also cause typechecking failure.
- If sandboxing blocks cache access for an important lint command, rerun the same narrow command with escalation instead of broadening the lint scope.
- After fixing lint issues, rerun the same narrow lint command and then a targeted `go build` or `go test` for the touched package.
