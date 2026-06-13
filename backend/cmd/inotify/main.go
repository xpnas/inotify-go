package main

import (
	"log"

	"inotify/backend/internal/config"
	"inotify/backend/internal/database"
	"inotify/backend/internal/handlers"
	"inotify/backend/internal/sender"
)

func main() {
	cfg := config.Load()
	store, err := database.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	srv := &handlers.Server{
		Store:  store,
		Sender: sender.New(store),
		UIRoot: LoadUI(),
	}
	log.Printf("inotify listening on %s", cfg.Addr)
	if err := srv.Router().Run(cfg.Addr); err != nil {
		log.Fatal(err)
	}
}
