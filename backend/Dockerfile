FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/inotify ./cmd/inotify

FROM alpine:3.20

WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=build /out/inotify /app/inotify
ENV INOTIFY_ADDR=:8000
ENV INOTIFY_DATA_DIR=/app/inotify_data
EXPOSE 8000
VOLUME ["/app/inotify_data"]
ENTRYPOINT ["/app/inotify"]
