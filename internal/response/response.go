package response

import (
	"encoding/json"
	"log"
	"net/http"
)

// Helper function to write JSON error responses
func WriteJSONError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	errorResponse := map[string]string{"error": message}
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		log.Printf("[ERROR] Failed to encode JSON error response: %v", err)
	}
}
