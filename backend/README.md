# Inotify Backend

Go API service for Inotify.

## Run

```bash
go run ./cmd/inotify
```

The service listens on `:8000` by default and stores runtime data in `inotify_data/`.

## Environment

- `INOTIFY_ADDR`: listen address, default `:8000`
- `INOTIFY_DATA_DIR`: data directory, default `inotify_data`

## Test

```bash
go test ./...
```

## Docker

```bash
docker build -t inotify-backend .
docker run --rm -p 8000:8000 -v inotify_data:/app/inotify_data inotify-backend
```
