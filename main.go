package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "modernc.org/sqlite"

	"github.com/dog4ik/stbl/api"
	"github.com/dog4ik/stbl/db"
	"github.com/dog4ik/stbl/utils"
	"github.com/joho/godotenv"
)

const PORT uint16 = 3030

//go:embed schema.sql
var ddl string

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("WARN: Error loading .env file\n")
	}

	ctx := context.Background()

	database_path := utils.ExpectEnv("DATABASE_PATH")

	env_port := utils.ExpectEnv("PORT")
	port, err := strconv.Atoi(env_port)
	if err != nil {
		log.Fatalf("Failed to convert env port to number")
	}

	conn, err := sql.Open("sqlite", fmt.Sprintf("%s?cache=shared", database_path))
	conn.SetMaxOpenConns(1)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %s", err)
	}
	defer conn.Close()
	log.Printf("%s", ddl)
	if _, err := conn.ExecContext(ctx, ddl); err != nil {
		log.Fatalf("Failet to run init migration: %s", err)
	}

	queries := db.New(conn)

	mux := http.NewServeMux()

	state := api.NewState(queries)

	mux.HandleFunc("POST /payout", state.PayoutHandler)
	mux.HandleFunc("POST /pay", state.PaymentHandler)
	mux.HandleFunc("POST /status", state.StatusHandler)
	mux.HandleFunc("POST /callback/pay", state.PaymentCallbackHandler)
	mux.HandleFunc("POST /callback/payout", state.PayoutCallbackHandler)

	log.Printf("Started Listening on port %d", port)

	err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), mux)
	log.Fatalf("Failed to listen and serve: %s", err)
}
