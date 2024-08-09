package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
	"github.com/operas-logicas/lenslocked/models"
)

type PostgresConfig struct {
	Host string
	Port string
	User string
	Password string
	DBName string
	SSLMode string
}

func (cfg PostgresConfig) String() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
}

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// Postgres config
	cfg := PostgresConfig{
		Host: os.Getenv("POSTGRES_HOST"),
		Port: os.Getenv("POSTGRES_PORT"),
		User: os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName: os.Getenv("POSTGRES_DBNAME"),
		SSLMode: os.Getenv("POSTGRES_SSLMODE"),
	}
	
	// Open db connection
	db, err := sql.Open("pgx", cfg.String())
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
