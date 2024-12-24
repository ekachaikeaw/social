package main

import (
	"log"

	"github.com/ekachaikeaw/social/internal/db"
	"github.com/ekachaikeaw/social/internal/env"
	"github.com/ekachaikeaw/social/internal/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "postgres://myuser:mypassword@localhost/mydatabase?sslmode=disable")

	conn, err := db.New(addr, 3, 3, "15m")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	store := store.NewStorage(conn)
	db.Seed(store, conn)
}