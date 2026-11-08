package main

import (
	"encoding/json"
	"foodwise/internal/recommender"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/api/recommend", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if r.Method != "POST" {
			http.Error(w, "Use POST", 405)
			return
		}
		var q recommender.Request
		if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
			http.Error(w, "Invalid JSON", 400)
			return
		}
		json.NewEncoder(w).Encode(recommender.Recommend(q))
	})
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Foodwise API listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
