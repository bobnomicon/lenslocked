package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
)

type PostgresConfig struct {
	Host string
	Port string
	User string
	Password string
	DBName string
	SSLMode string
}

var (
	flags = flag.NewFlagSet("goose", flag.ExitOnError)
	dir = flags.String("dir", "migrations", "Directory with migration files")
)

func (cfg PostgresConfig) String() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
}

func DefaultPostgresConfig() PostgresConfig {
	// Postgres config
	return PostgresConfig{
		Host: os.Getenv("POSTGRES_HOST"),
		Port: os.Getenv("POSTGRES_PORT"),
		User: os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName: os.Getenv("POSTGRES_DBNAME"),
		SSLMode: os.Getenv("POSTGRES_SSLMODE"),
	}
}

func main() {	
	// Parse flags
	flags.Parse(os.Args[1:])
	args := flags.Args()
	if len(args) == 0 {
		flags.Usage()
		return
	}

	// Get command and args
	command := args[0]
	arguments := []string{}
	if len(args) > 1 {
		arguments = append(arguments, args[1:]...)
	}

	// Load .env
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// Conect to DB
	cfg := DefaultPostgresConfig()
	db, err := goose.OpenDBWithDriver("postgres", cfg.String())
	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}

	// Close connection to DB on exit
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("goose: failed to close DB: %v\n", err)
		}
	}()

	// Run goose command
	ctx := context.Background()
	if err := goose.RunContext(ctx, command, db, *dir, arguments...); err != nil {
		log.Fatalf("goose %v: %v", command, err)
	}
}
