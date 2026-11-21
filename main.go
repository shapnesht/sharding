package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/buraksezer/consistent"
	"github.com/cespare/xxhash/v2"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type Response struct {
	Url      string `json:"url"`
	ShortUrl string `json:"shortUrl"`
	Server   string `json:"server"`
}

type hasher struct{}

func (h hasher) Sum64(data []byte) uint64 {
	return xxhash.Sum64(data)
}

type myMember string

func (m myMember) String() string {
	return string(m)
}

func handleGetUrl(w http.ResponseWriter, r *http.Request, c *consistent.Consistent) {
	fmt.Fprintf(w, "GET Hello, %q", r.URL.Path)
}

func handlePostUrl(w http.ResponseWriter, r *http.Request, c *consistent.Consistent) {
	url := r.URL.Query().Get("url")
	hash := sha256.Sum256([]byte(url))
	shortUrl := hex.EncodeToString(hash[:])[0:5]
	w.Header().Set("Content-Type", "application/json")
	server := c.LocateKey([]byte(shortUrl)).String()
	response := Response{
		ShortUrl: shortUrl,
		Url:      url,
		Server:   server,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, `{"error":"failed to encode JSON"}`, http.StatusInternalServerError)
	}
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
	fmt.Println("Starting the server....")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Unable to fetch configs")
	}

	host := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")

	fmt.Println("Connecting to DB....")
	conn1, err1 := connect(host, dbName, dbUsername, dbPassword, 5433)
	conn2, err2 := connect(host, dbName, dbUsername, dbPassword, 5434)
	conn3, err3 := connect(host, dbName, dbUsername, dbPassword, 5435)
	conn4, err4 := connect(host, dbName, dbUsername, dbPassword, 5436)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		log.Fatal("Unable to connect to DB...")
	}
	fmt.Println("Connected...")

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

	cfg := consistent.Config{
		PartitionCount:    7,
		ReplicationFactor: 1,
		Load:              1.25,
		Hasher:            hasher{},
	}
	c := consistent.New(nil, cfg)

	c.Add(myMember("5433"))
	c.Add(myMember("5434"))
	c.Add(myMember("5435"))
	c.Add(myMember("5436"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetUrl(w, r, c)
		case http.MethodPost:
			handlePostUrl(w, r, c)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
