# AGENTS

This file helps coding agents become productive quickly in this repository.

## Repository Overview

- Backend API: [backend](backend/)
- Frontend admin UI: [frontend](frontend/)
- Documentation: [docs](docs/)

The old .NET Core and Vue 2 code has been removed. Keep future work focused on the Go backend and Vue 3 frontend.

## Build, Run, and Test

Use the smallest relevant command for your change.

### Backend

- Test: `cd backend && go test ./...`
- Run locally: `cd backend && go run ./cmd/inotify`
- Build binary: `cd backend && go build ./cmd/inotify`
- Build Docker image: `docker build -t inotify-backend ./backend`

### Frontend

- Install deps: `cd frontend && npm install`
- Dev server: `cd frontend && npm run dev`
- Production build: `cd frontend && npm run build`
- Lint: `cd frontend && npm run lint`
- Build Docker image: `docker build -t inotify-frontend ./frontend`

### Full Stack

- Docker Compose: `docker compose up -d --build`
- Full validation:
  - `cd backend && go test ./...`
  - `cd frontend && npm run lint && npm run build`

## Architecture Boundaries

- API handlers are in [backend/internal/handlers](backend/internal/handlers/).
- Data access and migrations are in [backend/internal/database](backend/internal/database/).
- Data models are in [backend/internal/models](backend/internal/models/).
- Message channel implementations are in [backend/internal/sender](backend/internal/sender/).
- Frontend API wrappers are in [frontend/src/api](frontend/src/api/).
- Frontend pages are in [frontend/src/views](frontend/src/views/).

Keep changes scoped to one layer unless the task requires cross-layer updates.

## Project Conventions and Pitfalls

- Backend runtime data lives under `inotify_data/`; do not commit it.
- Default bootstrap account is `admin / 123456`; treat this as sensitive in deployment docs.
- The frontend dev server defaults to port `9000`.
- The backend dev server defaults to port `8000`.
- The frontend production Nginx proxies `/api`, Bark registration paths, health paths, and 32-hex send-key paths to the backend.

## Change Validation Expectations

- Backend-only change: run `go test ./...` in `backend`.
- Frontend-only change: run frontend lint and build.
- Full-stack change: run backend tests, frontend lint, and frontend build.

## Documentation Policy

- Prefer linking instead of duplicating long instructions.
- For deployment and API setup details, update files under [docs](docs/).

## context-mode Routing

context-mode MCP tools are available for this repository. Use them to keep large command output, file reads, web pages, and session recovery data out of the main conversation context.

### Mandatory Use Cases

- For analyzing, counting, filtering, comparing, searching, parsing, or transforming data, write code with `ctx_execute(language, code)` and print only the final answer.
- For multi-command repository discovery, use `ctx_batch_execute(commands, queries)` so raw output is indexed and only relevant snippets are returned.
- For follow-up questions over previously indexed output, use `ctx_search(queries: [...])`.
- For file analysis without editing, use `ctx_execute_file(path, language, code)`.
- For web pages, use `ctx_fetch_and_index(url, source)` and then `ctx_search(queries: [...])`.
- After a resume or compaction, search prior context before asking the user to restate work.

### Avoid

- Do not use `curl` or `wget` directly from shell for web content.
- Do not run inline HTTP fetches such as `node -e "fetch(...)"` or `python -c "requests.get(...)"`.
- Do not paste large command outputs, logs, raw HTML, or long file contents into the conversation.

### Command Selection

- Shell is fine for normal development commands such as `git`, `go test`, `npm install`, `npm run lint`, `npm run build`, and small `rg`/directory checks.
- If a shell command is likely to produce more than about 20 lines of output, route it through context-mode.
- Keep build, lint, and test commands at concurrency 1 because they share repository state and lock files.
