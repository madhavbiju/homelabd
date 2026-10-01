package response

import (
	"encoding/json"
	"net/http"

	"github.com/madhavbiju/homelabd/internal/ctxutil"
)

type APIError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}

// WriteError sends a standardized JSON error response.
func WriteError(w http.ResponseWriter, r *http.Request, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	reqID := ctxutil.GetRequestID(r.Context())

	resp := ErrorResponse{
		Error: APIError{
			Code:      code,
			Message:   message,
			RequestID: reqID,
		},
	}

	json.NewEncoder(w).Encode(resp)
}

// WriteJSON sends a standardized JSON response.
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
