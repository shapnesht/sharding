package main

import (
	"fmt"
	"log"
	"net/http"
)

func handleGetUrl(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "GET Hello, %q", r.URL.Path)
}

func handlePostUrl(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "POST Hello, %q", r.URL.Path)
}

func main() {
	fmt.Println("Running")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handleGetUrl(w, r)
		case "POST":
			handlePostUrl(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
