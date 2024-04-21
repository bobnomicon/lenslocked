package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	
	// Open db connection
	db, err := sql.Open("pgx", os.Getenv("DATA_SOURCE_NAME"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Verify db connection
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected!")
}
