# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AgentMesh is a multi-tenant AI Code Agent collaboration platform supporting Claude Code, Codex CLI, Gemini CLI, Aider, and more. It consists of three main components:

- **Backend**: Go API server (Gin + GORM + gqlgen)
- **Web**: Next.js frontend (App Router + TypeScript + Tailwind CSS)
- **Runner**: Go daemon that executes AI agent tasks in isolated PTY environments

## Development Environment (Docker)

**Always use `deploy/dev` Docker environment for development and debugging.** This setup includes Nginx reverse proxy and mirrors production architecture, helping catch issues early.

### Start Development Environment

```bash
cd deploy/dev
docker compose up -d             # Start all services
docker compose logs -f           # View logs
docker compose down              # Stop all services
docker compose down -v           # Stop and remove volumes
```

### Services & Ports

| Service | URL | Description |
|---------|-----|-------------|
| **App (Nginx)** | http://localhost | Unified entry point (proxies to web/backend) |
| **Adminer** | http://localhost:8081 | Database management UI |
| **MinIO Console** | http://localhost:9001 | S3-compatible storage UI |
| PostgreSQL | localhost:5432 | Database (user: agentmesh, pass: agentmesh_dev) |
| Redis | localhost:6379 | Cache |
| MinIO API | localhost:9000 | S3 API endpoint |

### Hot Reload

All services support hot reload - source code is mounted into containers:
- **Backend**: Go code changes auto-rebuild via Air
- **Web**: Next.js fast refresh
- **Runner**: Go code changes auto-rebuild

### Environment Configuration

Copy and customize environment variables:
```bash
cd deploy/dev
cp .env.example .env             # Edit as needed
```

## Build Commands (for CI/testing outside Docker)

### Backend (Go)

```bash
cd backend
go build ./cmd/server            # Build binary
go test ./...                    # Run all tests
go test -v ./internal/service/... -run TestAuth  # Run specific test
```

### Web (Next.js)

```bash
cd web
pnpm install                     # Install dependencies
pnpm build                       # Production build
pnpm lint                        # ESLint
pnpm test                        # Run tests (Vitest)
pnpm test:run                    # Run tests once
pnpm test:coverage               # Test coverage
```

### Runner (Go)

```bash
cd runner
make build                       # Build with desktop support (CGO)
make build-nocgo                 # Build CLI-only (no CGO)
make test                        # Run tests
make lint                        # golangci-lint
make build-all                   # Cross-platform builds
```

### Database Migrations

```bash
./scripts/migrate.sh up          # Apply migrations
./scripts/migrate.sh down 1      # Rollback last migration
./scripts/migrate.sh create xxx  # Create new migration
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Web (Next.js)                            │
│                 localhost:3000                              │
└─────────────────────────────────────────────────────────────┘
                              │
                    REST / GraphQL / WebSocket
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Backend (Go + Gin)                        │
│                 localhost:8080                              │
│  - Auth (JWT + OAuth)                                       │
│  - Organization/Team/User management                        │
│  - Pod lifecycle management                                 │
│  - Ticket/Channel management                                │
│  - PostgreSQL + Redis                                       │
└─────────────────────────────────────────────────────────────┘
                              │
                         WebSocket
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Runner (Go daemon)                        │
│              Self-hosted by users                           │
│  - Receives tasks via WebSocket                             │
│  - Creates isolated PTY terminals (Pods)                    │
│  - Executes AI agents (Claude Code, Aider, etc.)            │
│  - Streams terminal output back to server                   │
└─────────────────────────────────────────────────────────────┘
```

## Backend Structure

```
backend/
├── cmd/server/           # Entry point
├── internal/
│   ├── api/              # REST & GraphQL handlers
│   │   ├── rest/         # REST endpoints
│   │   └── graphql/      # GraphQL resolvers
│   ├── domain/           # Domain models (DDD style)
│   │   ├── user/         # User entity
│   │   ├── organization/ # Organization entity
│   │   ├── agentpod/     # AgentPod entity
│   │   ├── ticket/       # Ticket entity
│   │   ├── channel/      # Channel entity
│   │   └── runner/       # Runner entity
│   ├── service/          # Business logic layer
│   ├── infra/            # Infrastructure (DB, cache)
│   └── middleware/       # Auth, tenant isolation
├── migrations/           # SQL migrations
└── generated/            # gqlgen generated code
```

## Web Structure

```
web/src/
├── app/                  # Next.js App Router
│   ├── (auth)/           # Auth pages (login, register)
│   ├── (dashboard)/      # Dashboard pages
│   └── api/              # API routes
├── components/           # React components
├── lib/                  # Utilities, API clients
├── stores/               # Zustand state stores
├── hooks/                # Custom React hooks
├── messages/             # i18n translations (next-intl)
└── providers/            # Context providers
```

## Runner Structure

```
runner/
├── cmd/runner/           # Entry point (register/run/service/desktop)
├── internal/
│   ├── runner/           # Core runner logic
│   │   ├── runner.go     # Main Runner struct
│   │   ├── pod_handler.go    # Pod command processing
│   │   ├── pod_builder.go    # Builder pattern for Pods
│   │   └── message_handler.go # WebSocket message routing
│   ├── client/           # WebSocket client
│   │   ├── connection.go # Auto-reconnect logic
│   │   └── protocol.go   # Message types
│   ├── terminal/         # PTY management (creack/pty)
│   ├── sandbox/          # Sandbox environment
│   │   └── plugins/      # worktree, tempdir plugins
│   ├── luaplugin/        # Lua plugin system
│   ├── mcp/              # Model Context Protocol
│   └── workspace/        # Git worktree management
```

## Key Concepts

**Pod**: An isolated execution environment with PTY terminal, sandbox config, and output forwarder.

**Runner**: Self-hosted daemon that connects to backend via WebSocket, receives tasks, and manages Pod lifecycle.

**Sandbox**: Configurable environment created by plugins (worktree for Git isolation, tempdir for temporary workspace).

**Channel**: Multi-agent collaboration space where agents can communicate.

**Ticket**: Task management unit with kanban board integration.

## Message Flow (Runner ↔ Backend)

1. Backend sends `create_pod` → Runner creates Sandbox → Starts PTY
2. Terminal output → PTYForwarder → `terminal_output` to backend → WebSocket to web
3. User input from web → `terminal_input` → Runner writes to PTY stdin
4. Backend sends `terminate_pod` → Runner stops PTY → Cleans up Sandbox

## Configuration

**Backend**: `backend/.env` (copy from `.env.example`)
**Web**: `web/.env.local` (copy from `.env.example`)
**Runner**: `~/.agentsmesh/config.yaml` (created after `runner register`)

## GraphQL

GraphQL schema: `backend/internal/api/graphql/schema/*.graphqls`
Generate code: `cd backend && go generate ./...`

## Testing Patterns

- Backend: Standard Go testing with `testify`
- Web: Vitest + Testing Library
- Runner: Go testing, files ending with `_integration_test.go` for integration tests
