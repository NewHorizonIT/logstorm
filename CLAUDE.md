# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is LogStorm

LogStorm is a developer-first observability platform for log management. Users register, create workspaces, generate API keys, and send logs via the Go SDK. Logs are stored in ClickHouse and searched through the web UI.

## Repository Layout

```
app/api/          - Go REST API (Gin, PostgreSQL + ClickHouse)
app/web/          - Next.js 16.2 frontend (React 19, Tailwind v4, shadcn/ui)
packages/sdk-go/  - Go SDK for log ingestion (module: github.com/NewHorizonIT/logstorm-sdk-go)
docs/decisions/   - Architectural decision records
docs/product/     - Product requirements
deploy/           - Docker init scripts
```

## Commands

### Infrastructure

```bash
make up          # Start Docker services (postgres, clickhouse, adminer, mailpit)
make down        # Stop Docker services
make restart     # Restart Docker services
```

### Go API (`app/api/`)

```bash
make build       # Build binary to app/api/bin/logstorm
make run         # Build and run the server
make test        # Run all Go tests with verbose output

# Single package
cd app/api && go test ./internal/config/... -v

# Database migrations (requires golang-migrate CLI)
make migrate-up
make migrate-down
make migrate-create name=<migration_name>
```

### Next.js Web (`app/web/`)

```bash
npm run dev           # Start dev server (runs app/web next dev)
npm run lint          # ESLint
npm run typecheck:web # TypeScript check (no emit)
npm run format        # Prettier format all files
npm run format:check  # Check formatting without writing
```

## Architecture

### Go API

**Entry point:** `app/api/cmd/server/main.go` bootstraps `internal/bootstrap/NewApp()`, which wires config → logger → PostgreSQL → ClickHouse → Gin router, then starts an HTTP server with graceful shutdown.

**Module pattern:** Each domain lives under `internal/modules/<name>/` with its own `router.go` that registers routes on the shared `*gin.RouterGroup`. New domains follow this same pattern. Register new modules in `internal/bootstrap/router.go`.

**Config:** YAML file at `app/api/configs/config.yaml`, loaded by Viper. All sections have validation tags (go-playground/validator). The Go module path is `github.com/logstorm/api`.

**Logging:** Zerolog (`github.com/rs/zerolog`) with request middleware in `internal/logger/middleware.go`. Log rotation via lumberjack.

**Databases:**

- PostgreSQL (pgx/v5) — user data, workspaces, API keys; connection pooling configured in config
- ClickHouse (clickhouse-go/v2) — log event storage and analytics; uses LZ4 compression

**Migrations:** SQL files in `app/api/migrations/`, named `NNNNNN_<name>.up.sql` / `.down.sql`. Uses [golang-migrate](https://github.com/golang-migrate/migrate) CLI.

**API base path:** `/api/v1` (configured via `server.base_path` in config.yaml)

### Next.js Frontend

**Framework:** Next.js 16.2 with App Router. **This version has breaking changes vs earlier Next.js** — check `node_modules/next/dist/docs/` before writing Next.js-specific code. The AGENTS.md note in `app/web/` applies here.

**UI stack:** React 19 + Tailwind CSS v4 + shadcn/ui (built on Base UI primitives) + Phosphor Icons. Shared components live in `app/web/components/ui/`.

**Path alias:** `@/` resolves to the root of `app/web/` (e.g., `@/lib/utils`, `@/components/ui/button`).

**Theme:** `next-themes` with dark/light/system support, wrapped in `app/web/providers/theme-provider.tsx`.

**Fonts:** Geist Sans, Geist Mono, and JetBrains Mono (loaded via `next/font/google`).

### Data Flow

```
App (Go SDK) → POST /api/v1/logs → API (validates API key) → ClickHouse
User browser  → GET  /api/v1/...  → API (JWT auth)          → PostgreSQL / ClickHouse
```

## Git Conventions

Commits must follow [Conventional Commits](https://www.conventionalcommits.org/) — enforced via `commitlint` on the `commit-msg` hook. Common types: `feat`, `fix`, `refactor`, `docs`, `chore`.

`lint-staged` runs ESLint + Prettier on staged files before each commit.

## Infrastructure Services (Docker)

| Service    | Image                   | Default Port                                                       |
| ---------- | ----------------------- | ------------------------------------------------------------------ |
| PostgreSQL | postgres:16             | `$POSTGRES_PORT`                                                   |
| ClickHouse | clickhouse-server:24.8  | `$CLICKHOUSE_HTTP_PORT` (HTTP), `$CLICKHOUSE_NATIVE_PORT` (native) |
| Adminer    | adminer:5.4.2           | `$ADMINER_PORT`                                                    |
| Mailpit    | axllent/mailpit:v1.30.1 | `$MAILPIT_SMTP_PORT` / `$MAILPIT_WEB_PORT`                         |

Copy `.env.example` to `.env` and fill in port values before running `make up`.
