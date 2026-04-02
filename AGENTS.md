# Repo Guidance

## Go linting

- Run `golangci-lint` per directory or file in this repo, not `./...` from the module root.
- Prefer narrow commands for only the paths you changed, for example:
  - `cd apps/api/handlers && golangci-lint run`
  - `cd apps/api && golangci-lint run main.go`
- The Go workspace layout and sandbox/cache behavior can make broad package-loading unreliable, so root/module-wide lint runs may fail before producing useful findings.
- Treat cache persistence warnings as non-blocking noise unless they also cause typechecking failure.
- If sandboxing blocks cache access for an important lint command, rerun the same narrow command with escalation instead of broadening the lint scope.
- After fixing lint issues, rerun the same narrow lint command and then a targeted `go build` or `go test` for the touched package.
