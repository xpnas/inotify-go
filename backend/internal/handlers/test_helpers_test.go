package handlers

import (
	"testing"

	"inotify/backend/internal/config"
	"inotify/backend/internal/database"
)

func openTestStore(t *testing.T) *database.Store {
	t.Helper()
	root := t.TempDir()
	store, err := database.Open(config.Config{
		Addr:    "127.0.0.1:0",
		DataDir: root,
		DBPath:  root + "/test.db",
		JWTPath: root + "/jwt.json",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if sqlDB, err := store.DB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return store
}
