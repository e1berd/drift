package main

import (
	"context"
	"log"
	"net/url"
	"os"

	"github.com/e1berd/drift/app/internal/database"
)

func main() {
	ctx := context.Background()

	dsn, err := url.Parse(os.ExpandEnv(os.Getenv("DATABASE_URL")))
	if err != nil {
		log.Fatal("invalid DATABASE_URL: ", err)
	}

	pool, err := database.Connect(ctx, dsn)

	if err != nil {
		log.Fatal("database connection failed: ", err)
	}

	defer pool.Close()
}
