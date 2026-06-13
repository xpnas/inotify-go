# Stage 1: Build frontend
FROM node:22-alpine AS frontend-build
WORKDIR /src
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Build backend with embedded frontend
FROM golang:1.22-alpine AS backend-build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
# Copy frontend dist into the embed path expected by cmd/inotify/ui.go
COPY --from=frontend-build /src/dist ./cmd/inotify/ui/dist
RUN CGO_ENABLED=0 go build -o /out/inotify ./cmd/inotify

# Stage 3: Minimal runtime image
FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=backend-build /out/inotify /app/inotify
ENV INOTIFY_ADDR=:8000
ENV INOTIFY_DATA_DIR=/app/inotify_data
EXPOSE 8000
VOLUME ["/app/inotify_data"]
ENTRYPOINT ["/app/inotify"]
