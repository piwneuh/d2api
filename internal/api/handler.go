package api

import (
	"fmt"
	"net/http"
)

type Handler struct{}

// NewHandler creates a new instance of the API handler.
func NewHandler() *Handler {
	return &Handler{}
}

// HandleHello handles requests to the / endpoint.
func (h *Handler) HandleHello(w http.ResponseWriter, r *http.Request) {
	// Write "Hello, World!" to the response.
	fmt.Fprintf(w, "Hello, World!")
}
