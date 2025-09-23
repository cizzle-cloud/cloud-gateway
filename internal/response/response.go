package response

import (
	"fmt"
	"log"
	"net/http"
)

// Helper function to write JSON error responses
func WriteJSONError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, message)))
	if err != nil {
		log.Printf("[ERROR] Failed to encode JSON error response: %v", err)
	}
}
