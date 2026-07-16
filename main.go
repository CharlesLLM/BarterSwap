package main

import (
	"context"
	"log/slog"
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
		slog.Error("la variable DATABASE_URL est obligatoire")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store, err := postgres.New(ctx, databaseURL)
	if err != nil {
		slog.Error("initialisation du store postgres", "error", err)
		return
	}
	defer func() {
		if err := store.Close(); err != nil {
			slog.Error("fermeture de la base de données", "error", err)
		}
	}()

	if err := store.CreateSchema(ctx); err != nil {
		slog.Error("création du schéma SQL", "error", err)
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

	slog.Info("serveur démarré", "url", "http://localhost:8080")

	if err := server.ListenAndServe(); err != nil {
		slog.Error("arrêt du serveur HTTP", "error", err)
	}
}
