package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/CharlesLLM/BarterSwap/internal/application"
	"github.com/CharlesLLM/BarterSwap/internal/httpapi"
	"github.com/CharlesLLM/BarterSwap/internal/postgres"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("la variable DATABASE_URL est obligatoire")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store, err := postgres.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("fermeture de la base de données : %v", err)
		}
	}()

	if err := store.CreateSchema(ctx); err != nil {
		log.Fatal(err)
	}

	userService := application.NewUserService(store)
	serviceService := application.NewServiceService(store)
	handler := httpapi.NewHandler(userService, serviceService)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("serveur démarré sur http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
