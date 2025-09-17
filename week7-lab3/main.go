package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var db *sql.DB

func initDB() {
	host := getEnv("DB_HOST", "localhost")
	name := getEnv("DB_NAME", "bookstore")
	user := getEnv("DB_USER", "bookstore_user")
	password := getEnv("DB_PASSWORD", "your_password")
	port := getEnv("DB_PORT", "5432")

	conSt := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, name)

	var err error
	db, err = sql.Open("postgres", conSt)
	if err != nil {
		log.Fatal("failed to open database: ", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	log.Println("successfully connected to database")
}

func main() {
	initDB()
}