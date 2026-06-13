# Migration

This repository was reorganized into a new Go + Vue 3 project.

## Removed

- `Inotify/`: old .NET Core backend
- `Inotify.Vue/`: old Vue 2 admin UI
- `sonar.sln`: old solution file
- `public/`: old screenshot assets

## Current Layout

- `backend/`: Go API server
- `frontend/`: Vue 3 + Vite admin UI
- `docs/`: project documentation
- `.github/workflows/ci.yml`: validation and Docker automation

## Data

The Go backend uses SQLite and auto-migrates tables on startup. Runtime data defaults to `backend/inotify_data` in local development and `/app/inotify_data` in Docker.
