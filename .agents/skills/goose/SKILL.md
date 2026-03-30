---
name: goose
description: Expert guide for working with goose, the Go database migration tool (github.com/pressly/goose). Use this skill whenever the user asks about database migrations using goose, including creating migration files, running up/down migrations, writing SQL or Go migrations, embedding migrations, configuring goose via environment variables, handling out-of-order migrations, using the goose CLI, or integrating goose as a Go library. Trigger on phrases like "goose migrate", "create a migration", "run migrations", "goose up/down", "SQL migration file", "Go migration function", "goose status", "rollback migration", "goose fix", or any question about managing database schema changes with goose.
---

# Goose Database Migration Tool

Goose (github.com/pressly/goose) manages database schemas via incremental SQL or Go function migrations. It supports Postgres, MySQL, SQLite, MSSQL, ClickHouse, Spanner, YDB, and more.

## Install

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
# or macOS:
brew install goose
```

## CLI Usage Pattern

```
goose DRIVER DBSTRING [OPTIONS] COMMAND
# or via env vars:
GOOSE_DRIVER=postgres GOOSE_DBSTRING="..." goose COMMAND
```

**Supported drivers:** `postgres`, `mysql`, `sqlite3`, `mssql`, `spanner`, `redshift`, `tidb`, `clickhouse`, `ydb`, `starrocks`, `turso`

### Common Connection Strings

| Driver | Example |
|--------|---------|
| postgres | `"user=postgres dbname=mydb sslmode=disable"` |
| mysql | `"user:password@/dbname?parseTime=true"` |
| sqlite3 | `./foo.db` |
| mssql | `"sqlserver://user:pass@host:1433?database=mydb"` |
| clickhouse | `"tcp://127.0.0.1:9000"` |

---

## Core Commands

| Command | Description |
|---------|-------------|
| `up` | Apply all pending migrations |
| `up-by-one` | Apply exactly one migration |
| `up-to VERSION` | Migrate to a specific version |
| `down` | Roll back the latest migration |
| `down-to VERSION` | Roll back to a specific version |
| `down-to 0` | Roll back ALL migrations (destructive!) |
| `reset` | Roll back all migrations |
| `redo` | Re-run the latest migration |
| `status` | Show applied/pending migration status |
| `version` | Print current DB version |
| `create NAME [sql\|go]` | Create a new migration file |
| `fix` | Renumber timestamped → sequential |
| `validate` | Check migration files without running them |

### Examples

```bash
# Apply all pending migrations
goose postgres "user=postgres dbname=mydb sslmode=disable" up

# Check status
goose sqlite3 ./foo.db status

# Create a new SQL migration
goose sqlite3 ./foo.db create add_users_table sql
# → 20240315142300_add_users_table.sql

# Create with sequential numbering (-s flag)
goose -s sqlite3 ./foo.db create add_users_table sql
# → 00001_add_users_table.sql

# Roll back one migration
goose postgres "..." down

# Use a custom migrations directory
goose -dir ./db/migrations postgres "..." up
```

---

## Environment Variables

Avoid passing credentials on the command line:

```bash
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="postgres://admin:admin@localhost:5432/mydb"
export GOOSE_MIGRATION_DIR=./migrations
export GOOSE_TABLE=custom_schema.goose_migrations  # optional custom table
```

Or use a `.env` file in the project root:

```
GOOSE_DRIVER=postgres
GOOSE_DBSTRING=postgres://admin:admin@localhost:5432/mydb
GOOSE_MIGRATION_DIR=./migrations
```

Disable `.env` loading with `-env=none`. Point to a specific file with `-env=./path/.env`.

---

## SQL Migrations

### Basic Structure

```sql
-- +goose Up
CREATE TABLE users (
    id   SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE
);

-- +goose Down
DROP TABLE users;
```

**Rules:**
- Every file **must** have `-- +goose Up`
- `-- +goose Down` is optional but recommended
- `Up` annotation must come before `Down`
- Statements must end with `;`
- Migrations run inside a transaction by default

### Skip Transactions (DDL that can't run in transactions)

```sql
-- +goose NO TRANSACTION

-- +goose Up
CREATE DATABASE mydb;

-- +goose Down
DROP DATABASE mydb;
```

### Complex Statements (PL/pgSQL, stored procedures)

Use `StatementBegin`/`StatementEnd` for multi-line statements with internal semicolons:

```sql
-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION IF EXISTS update_timestamp();
```

### Environment Variable Substitution

```sql
-- +goose Up
-- +goose ENVSUB ON
INSERT INTO config (key, value) VALUES ('app_env', '${APP_ENV}');
INSERT INTO config (key, value) VALUES ('region', '${REGION:-us-east-1}');
-- +goose ENVSUB OFF
```

**Supported expansion syntax:**
- `${VAR}` or `$VAR` — value of VAR
- `${VAR:-default}` — value of VAR, or `default` if unset/null
- `${VAR-default}` — value of VAR, or `default` if unset
- `${VAR?err_msg}` — value of VAR, or error with `err_msg` if unset

---

## Go Migrations

Use when schema changes require application logic (e.g., data transformation).

### Step 1: Create your own goose binary

```
myapp/
├── main.go
└── migrations/
    ├── 00001_create_users.sql
    └── 00002_backfill_emails.go
```

### Step 2: Write migration functions

```go
// migrations/00002_backfill_emails.go
package migrations

import (
    "database/sql"
    "github.com/pressly/goose/v3"
)

func init() {
    goose.AddMigration(upBackfillEmails, downBackfillEmails)
}

func upBackfillEmails(tx *sql.Tx) error {
    _, err := tx.Exec(`UPDATE users SET email = lower(email) WHERE email IS NOT NULL`)
    return err
}

func downBackfillEmails(tx *sql.Tx) error {
    // Rollback logic (if reversible)
    return nil
}
```

### Step 3: Import migrations in main.go

```go
package main

import (
    "database/sql"
    "log"

    "github.com/pressly/goose/v3"
    _ "github.com/lib/pq"
    _ "myapp/migrations" // side-effect import registers migrations
)

func main() {
    db, err := sql.Open("postgres", "user=postgres dbname=mydb sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    if err := goose.Up(db, "migrations"); err != nil {
        log.Fatal(err)
    }
}
```

**Go migration file rules:**
- Filename must start with a number followed by `_`
- Must not end with `_test.go`
- Register via `goose.AddMigration(upFn, downFn)` in `init()`

---

## Embedded Migrations (Go 1.16+)

Bundle migration files into the compiled binary:

```go
package main

import (
    "database/sql"
    "embed"

    "github.com/pressly/goose/v3"
    _ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
    db, _ := sql.Open("postgres", "...")

    goose.SetBaseFS(embedMigrations)

    if err := goose.SetDialect("postgres"); err != nil {
        panic(err)
    }
    if err := goose.Up(db, "migrations"); err != nil {
        panic(err)
    }
}
```

Note: `fix` and `create` commands still operate on the OS filesystem — `embed.FS` is read-only.

---

## Out-of-Order Migrations

By default, goose errors if you try to apply a migration with a version older than the current DB version (common in team environments when branches diverge).

```bash
# Allow applying missing/out-of-order migrations
goose -allow-missing postgres "..." up
```

As a library:
```go
goose.Up(db, "migrations", goose.WithAllowMissing())
```

**Recommended team workflow (hybrid versioning):**
1. Developers create timestamped migrations locally (`20240315_add_column.sql`)
2. Before deploying to production, run `goose fix` to renumber them sequentially
3. Run `fix` in CI pipelines before production deploys

```bash
goose fix  # converts timestamps → sequential numbers, preserves order
```

---

## CLI Flags Reference

| Flag | Description |
|------|-------------|
| `-dir PATH` | Migration directory (default: `.`) |
| `-s` | Use sequential numbers instead of timestamps for `create` |
| `-v` | Verbose output |
| `-allow-missing` | Apply out-of-order migrations |
| `-no-versioning` | Run migrations in file order without tracking versions |
| `-table NAME` | Custom version table name (default: `goose_db_version`) |
| `-timeout DURATION` | Max query duration (e.g., `30s`, `5m`) |
| `-no-color` | Disable colored output |
| `-env PATH` | Path to `.env` file; use `-env=none` to disable |

---

## MySQL-Specific Notes

- Add `?parseTime=true` to the connection string for `status` to work correctly
- Add `?multiStatements=true` for SQL files with multiple statements separated by `;`

Example: `"user:password@/dbname?parseTime=true&multiStatements=true"`

---

## Custom Migrations Table / Schema

```bash
# Use a non-public schema
goose -table myschema.goose_db_version postgres "..." up
```

Or via environment variable:
```bash
export GOOSE_TABLE=myschema.goose_db_version
```

---

## Slim Binary Build (exclude unneeded drivers)

```bash
go build -tags='no_postgres no_mysql no_sqlite3 no_ydb' -o goose ./cmd/goose
# Available tags: no_clickhouse no_libsql no_mssql no_mysql
#                 no_postgres no_sqlite3 no_vertica no_ydb
```

---

## Quick-Reference Checklist

When helping with goose tasks:

1. **Creating a migration** → use `goose create NAME sql` (or `go`), then fill in `Up`/`Down` blocks
2. **Running migrations** → confirm driver and connection string, use `goose up`
3. **Rolling back** → `goose down` (one step) or `goose down-to VERSION`
4. **Checking state** → `goose status` and `goose version`
5. **Complex SQL** (stored procs, PL/pgSQL) → wrap with `StatementBegin`/`StatementEnd`
6. **No-transaction DDL** → add `-- +goose NO TRANSACTION` at top of file
7. **Team conflicts** → use `goose fix` before deploying; `-allow-missing` for local dev
8. **Embedding in binary** → use `//go:embed` + `goose.SetBaseFS()`