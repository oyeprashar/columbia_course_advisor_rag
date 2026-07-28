package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
)

// TODO : Bad code! Supplying credentials
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {

	// TODO : Bad code! Supplying credentials
	connStr := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		getEnv("PGHOST", "localhost"),
		getEnv("PGPORT", "5432"),
		getEnv("PGDATABASE", "course_advisor"),
		getEnv("PGUSER", "myuser"),
		getEnv("PGPASSWORD", "mypassword"),
	)

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	// 146 rows
	var courseCount int
	if err := db.QueryRow("SELECT count(*) FROM courses").Scan(&courseCount); err != nil {
		log.Fatalf("courses query failed: %v", err)
	}
	fmt.Printf("courses table: %d rows\n", courseCount)
}
