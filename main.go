package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("la variable DATABASE_URL est obligatoire")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store, err := NewStore(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}

	defer store.Close()

	if err := store.CreateSchema(ctx); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/users", usersHandler(store))
	mux.HandleFunc("/api/users/", userHandler(store))
	mux.HandleFunc("/api/services", servicesHandler(store))
	mux.HandleFunc("/api/services/", serviceHandler(store))

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("serveur démarré sur http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
