package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Middleware Handling CORS unblocking response
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Initalize Database
	fmt.Println("Initalize Database")
	db = initDB()
	defer db.Close()

	content, err := os.ReadFile("seed-data.json")
	if err != nil {
		log.Printf("Could not read seed-data.json: %v", err)
	} else {
		// Seed the Database
		err = SeedDatabase(string(content))
		if err != nil {
			log.Printf("Seeding failed: %v", err)
		}
	}

	// Set the Mux
	mux := http.NewServeMux()
	mux.HandleFunc("/api/payments", paymentHandler)
	mux.HandleFunc("/api/invoices/", InvoicesDispatcher)

	// Server Configuration
	server := &http.Server{
		Addr:         ":8080",
		Handler:      enableCORS(mux), // Wrap with CORS middleware
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("Server listening on http://localhost:8080")
	// Start Server
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("Server failed: ", err)
	}
}
