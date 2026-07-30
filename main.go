package main

import (
	"github.com/oyeprashar/columbia_course_advisor_rag/api"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/oyeprashar/columbia_course_advisor_rag/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("no .env file found, reading from actual environment variables")
	}

	err := database.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", api.HandleHealth)
	mux.HandleFunc("/recommend", api.HandleRecommend)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
