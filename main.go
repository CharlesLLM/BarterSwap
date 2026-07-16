package main

import (
	"context"
	"fmt"
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
		fmt.Println("la variable DATABASE_URL est obligatoire")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store, err := postgres.New(ctx, databaseURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err := store.Close(); err != nil {
			fmt.Printf("fermeture de la base de données : %v\n", err)
		}
	}()

	if err := store.CreateSchema(ctx); err != nil {
		fmt.Println(err)
		return
	}

	userService := application.NewUserService(store)
	catalogService := application.NewCatalogService(store)
	exchangeService := application.NewExchangeService(store)
	reviewService := application.NewReviewService(store)
	handler := httpapi.NewHandler(userService, catalogService, exchangeService, reviewService)

	server := http.Server{
		Addr:              ":8080",
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Println("serveur démarré sur http://localhost:8080")

	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
}
