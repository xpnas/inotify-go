# Development

## Requirements

- Go 1.22+
- Node.js 22+
- npm 10+
- Docker Desktop or Docker Engine

## Backend

```bash
cd backend
go test ./...
go run ./cmd/inotify
```

Environment:

- `INOTIFY_ADDR`: listen address, default `:8000`
- `INOTIFY_DATA_DIR`: runtime data directory, default `inotify_data`

## Frontend

```bash
cd frontend
npm install
npm run lint
npm run dev
```

Vite serves the admin UI on `http://localhost:9000` and proxies `/api` to `http://127.0.0.1:8000`.

## Full Validation

```bash
cd backend
go test ./...

cd ../frontend
npm run lint
npm run build
```

## Project Conventions

- Backend handlers live in `backend/internal/handlers`.
- Backend data models live in `backend/internal/models`.
- Message channel implementations live in `backend/internal/sender`.
- Frontend API wrappers live in `frontend/src/api`.
- Frontend pages live in `frontend/src/views`.
