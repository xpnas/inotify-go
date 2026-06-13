package config

import (
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Addr    string
	DataDir string
	DBPath  string
	JWTPath string
}

func Load() Config {
	dataDir := getenv("INOTIFY_DATA_DIR", "inotify_data")
	addr := getenv("INOTIFY_ADDR", ":8000")
	return Config{
		Addr:    addr,
		DataDir: dataDir,
		DBPath:  filepath.Join(dataDir, "inotify.db"),
		JWTPath: filepath.Join(dataDir, "jwt.json"),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func GetenvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if n, err := strconv.Atoi(value); err == nil {
			return n
		}
	}
	return fallback
}
