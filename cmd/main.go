package main

import (
	"d2api/internal/api"
	"fmt"
	"net/http"
)

func main() {
	// Initialize your API handler
	handler := api.NewHandler()

	// Define your HTTP routes
	http.HandleFunc("/", handler.HandleHello)

	// Start the HTTP server
	port := 8080
	fmt.Printf("Server is listening on :%d...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
