package main

import (
	"d2api/internal/api"
	"fmt"
	"net/http"
)

func main() {
	handler := api.NewHandler()

	http.HandleFunc("/", handler.HandleHello)

	port := 8080
	fmt.Printf("Server is listening on :%d...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
