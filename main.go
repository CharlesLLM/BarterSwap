package main

import (
	"database/sql"
	_ "embed"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

//go:embed schema.sql
var schema string

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://barterswap:barterswap@localhost:5432/barterswap?sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	if _, err = db.Exec(schema); err != nil {
		log.Fatal(err)
	}
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("BarterSwap écoute sur %s", addr)
	srv := &http.Server{
		Addr:              addr,
		Handler:           NewAPI(NewStore(db), log.Default()),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
