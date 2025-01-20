package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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

// Handler to fetch exchange rate for the given date and currency
func getExchangeRate(w http.ResponseWriter, r *http.Request) {
	// Rate limiting check
	if !limiter.Allow() {
		http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
		return
	}

	// Extract the date and currency code from the URL path
	parts := strings.Split(r.URL.Path[1:], "/")
	if len(parts) < 1 || len(parts) > 2 {
		http.Error(w, "Invalid URL format. Use /[date]/[currency] or /[date]", http.StatusBadRequest)
		return
	}

	date := parts[0]
	var currency string
	if len(parts) == 2 {
		currency = strings.ToUpper(parts[1]) // Capitalize currency code
	}

	// Fetch exchange rates for the requested date
	rateData, exists := exchangeRates[date]
	if !exists {
		http.Error(w, "Exchange rates not found for the specified date", http.StatusNotFound)
		return
	}

	// If currency is specified, return just that currency's exchange rate
	if currency != "" {
		rate, exists := rateData[currency]
		if !exists {
			http.Error(w, "Currency not found for the specified date", http.StatusNotFound)
			return
		}
		// Respond with the exchange rate for the specific currency
		w.Header().Set("Content-Type", "application/json")
		response := map[string]float64{
			currency: rate,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
		return
	}

	// If no currency is specified, return all the exchange rates for the date
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
