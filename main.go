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

func connect(host, dbName, username, password string, port int16) (*pgx.Conn, error) {
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", username, password, host, port, dbName)
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

	conn1, err1 := connect(host, dbName, dbUsername, dbPassword, 5433)
	conn2, err2 := connect(host, dbName, dbUsername, dbPassword, 5434)
	conn3, err3 := connect(host, dbName, dbUsername, dbPassword, 5435)
	conn4, err4 := connect(host, dbName, dbUsername, dbPassword, 5436)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		log.Fatal("Unable to connect to DB...")
	}

	connections := map[int]*pgx.Conn{
		5433: conn1,
		5434: conn2,
		5435: conn3,
		5436: conn4,
	}
	defer connections[5433].Close(context.Background())
	defer connections[5434].Close(context.Background())
	defer connections[5435].Close(context.Background())
	defer connections[5436].Close(context.Background())
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
