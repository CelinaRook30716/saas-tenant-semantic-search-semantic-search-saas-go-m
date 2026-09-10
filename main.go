package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	c, err := NewInfraiClient()
	if err != nil {
		log.Fatal(err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		q, tenant := r.URL.Query().Get("q"), r.URL.Query().Get("tenant")
		if q == "" || tenant == "" {
			http.Error(w, "q and tenant are required", 400)
			return
		}
		k, _ := strconv.Atoi(r.URL.Query().Get("top_k"))
		if k <= 0 {
			k = 5
		}
		hits, err := c.Search(q, tenant, k)
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"query": q, "tenant": tenant, "matches": hits})
	})
	log.Printf("tenant search listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
