package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/time/rate"
)

var exchangeRates map[string]map[string]float64
var limiter *rate.Limiter

// Load exchange rates from the JSON file
func loadExchangeRates() error {
	file, err := os.Open("exchange_rates.json")
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&exchangeRates)
}

// Handler to fetch exchange rate for the given date
func getExchangeRate(w http.ResponseWriter, r *http.Request) {
	// Rate limiting check
	if !limiter.Allow() {
		http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
		return
	}

	// Extract the date from the URL path
	date := r.URL.Path[len("/"):] // Strip the leading slash

	// Fetch exchange rate for the requested date
	rateData, exists := exchangeRates[date]
	if !exists {
		http.Error(w, "Exchange rates not found for the specified date", http.StatusNotFound)
		return
	}

	// Respond with JSON of the exchange rates for that date
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(rateData); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func main() {
	// Load exchange rates from JSON file
	if err := loadExchangeRates(); err != nil {
		fmt.Printf("Error loading exchange rates: %v\n", err)
		return
	}

	port := os.Getenv("PORT")
 
	if port == "" {
		port = "8080"
	}

	// Initialize rate limiter: Allow 5 requests per second, with a maximum burst of 10 requests
	limiter = rate.NewLimiter(rate.Every(time.Second), 10)

	// Set up HTTP server and routes
	http.HandleFunc("/", getExchangeRate)
	log.Println("Listening on port", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
