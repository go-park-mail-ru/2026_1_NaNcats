package response

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "apllication/json")
	w.WriteHeader(statusCode)

	if data == nil {
		data = map[string]string{}
	}

	if err := json.NewEncoder(w); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func Error(w http.ResponseWriter, statusCode int, message string) {
	JSON(w, statusCode, map[string]string{"error: ": message})
}
