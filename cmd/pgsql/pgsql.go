package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
)

type PostgresConfig struct {
	Host string
	Port string
	User string
	Password string
	DBName string
	SSLMode string
}

type Order struct {
	ID int
	UserID int
	Amount int
	Description string
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

	// !Create tables
	// _, err = db.Exec(`
	// 	CREATE TABLE IF NOT EXISTS users (
	// 		id SERIAL PRIMARY KEY,
	// 		name TEXT,
	// 		email TEXT NOT NULL
	// 	);

	// 	CREATE TABLE IF NOT EXISTS orders (
	// 		id SERIAL PRIMARY KEY,
	// 		user_id INT NOT NULL,
	// 		amount INT,
	// 		description TEXT
	// 	);
	// `)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Tables created.")

	// !Insert user
	// name := "Robert Miller"
	// email := "miller.robert.john@gmail.com"

	// _, err = db.Exec(`
	// 	INSERT INTO users (name, email)
	// 	VALUES ($1, $2);
	// `, name, email)

	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("User created.")

	// !Query single record
	var name, email string
	id := 1

	row := db.QueryRow(`
		SELECT name, email
		FROM users
		WHERE id=$1
	`, id)

	err = row.Scan(&name, &email)
	if err == sql.ErrNoRows {
		fmt.Println("Error, no rows!")
	} else if err != nil {
		panic(err)
	}
	fmt.Printf("User information: name=%s, email=%s\n", name, email)

	// !Insert fake orders
	// userID := 1

	// for i := 1; i <=5; i++ {
	// 	amount := i * 100
	// 	desc := fmt.Sprintf("Fake order #%d", i)

	// 	_, err = db.Exec(`
	// 		INSERT INTO orders (user_id, amount, description)
	// 		VALUES ($1, $2, $3)
	// 	`, userID, amount, desc)

	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Println("Created fake orders.")
	// }

	// !Query multiple records
	var orders []Order

	userID := 1
	rows, err := db.Query(`
		SELECT id, amount, description
		FROM orders
		WHERE user_id=$1
	`, userID)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var order Order
		order.UserID = userID
		err := rows.Scan(&order.ID, &order.Amount, &order.Description)
		if err != nil {
			panic(err)
		}
		orders = append(orders, order)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	fmt.Println("Orders:", orders)
}
