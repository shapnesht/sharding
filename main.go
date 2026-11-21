package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func handleGetUrl(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "GET Hello, %q", r.URL.Path)
}

func handlePostUrl(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "POST Hello, %q", r.URL.Path)
}

func connect(host, dbName, username, password string) (*pgx.Conn, error) {
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:5433/%s", username, password, host, dbName)
	conn, err := pgx.Connect(context.Background(), connectionString)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func main() {
	fmt.Println("Running")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Unable to fetch configs")
	}

	host := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")

	conn, err := connect(host, dbName, dbUsername, dbPassword)
	if err != nil {
		log.Fatal("Unable to connect to DB...")
	}
	defer conn.Close(context.Background())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetUrl(w, r)
		case http.MethodPost:
			handlePostUrl(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
