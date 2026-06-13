# Deployment

## Prerequisites

Install Docker and Docker Compose first:

- Windows/macOS: install Docker Desktop.
- Linux: install Docker Engine and Compose v2.

Verify:

```bash
docker --version
docker compose version
```

## Docker Compose

```bash
cp .env.example .env
docker compose up -d --build
```

Open `http://localhost:9000`.

Runtime data is stored in the `inotify_data` Docker volume.

Useful commands:

```bash
docker compose ps
docker compose logs -f
docker compose restart
docker compose down
```

## GitHub Container Registry

The workflow at `.github/workflows/ci.yml` builds and pushes images to GitHub Container Registry on pushes to `main` or `master`:

```text
ghcr.io/<owner>/<repo>/backend:latest
ghcr.io/<owner>/<repo>/frontend:latest
ghcr.io/<owner>/<repo>/backend:<commit-sha>
ghcr.io/<owner>/<repo>/frontend:<commit-sha>
```

For private repositories, sign in before pulling images:

```bash
echo <github-token> | docker login ghcr.io -u <github-username> --password-stdin
```

## Reverse Proxy

When exposing Inotify behind your own proxy, route traffic to the frontend container. The frontend Nginx container already proxies these backend paths:

- `/api/*`
- `/Register`
- `/RegisterCheck`
- `/Ping`
- `/Healthz`
- `/Info`
- `/<32-hex-send-key>/*`

## Backup

The SQLite database and JWT settings are stored in the backend data directory:

```text
inotify.db
jwt.json
```

For Docker Compose deployments:

```bash
docker run --rm -v inotify_inotify_data:/data -v "$PWD:/backup" alpine tar czf /backup/inotify-data.tgz -C /data .
```

## Upgrade

```bash
docker compose pull
docker compose up -d --build
```

The backend runs database auto-migrations at startup.
