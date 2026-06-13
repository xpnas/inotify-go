# Inotify Frontend

Vue 3 admin UI for Inotify.

## Commands

```bash
npm install
npm run dev
npm run lint
npm run build
```

The dev server listens on port `9000` and proxies `/api` to `http://127.0.0.1:8000` by default.

Override the proxy target:

```bash
set VITE_API_PROXY=http://127.0.0.1:18080
npm run dev
```

## Docker

```bash
docker build -t inotify-frontend .
```

The production image serves static files with Nginx and proxies backend paths to the Compose service named `backend`.
