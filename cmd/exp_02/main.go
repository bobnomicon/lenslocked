package main

import (
	"fmt"

	"github.com/operas-logicas/lenslocked/models"
)

func main() {
	// Open db connection
	cfg := models.DefaultPostgresConfig()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Verify db connection
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Database connected.")

	// Create user
	us := models.UserService{
		DB: db,
	}
	user, err := us.Create("bobert@bobmiller.com", "bob's secret 123")
	if err != nil {
		panic(err)
	}
	fmt.Println(user)
}
