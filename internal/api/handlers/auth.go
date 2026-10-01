package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/madhavbiju/homelabd/internal/api/response"
	"github.com/madhavbiju/homelabd/internal/ctxutil"
	"github.com/madhavbiju/homelabd/internal/audit"
	"github.com/madhavbiju/homelabd/internal/auth"
	"github.com/madhavbiju/homelabd/internal/database"
)

type AuthHandler struct {
	authService  *auth.Service
	auditService *audit.Service
	db           *database.Database
}

func NewAuthHandler(authService *auth.Service, auditService *audit.Service, db *database.Database) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		auditService: auditService,
		db:           db,
	}
}

type CreateTokenRequest struct {
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type CreateTokenResponse struct {
	Token string            `json:"token"`
	ID    string            `json:"id"`
	Perms []string          `json:"permissions"`
}

func (h *AuthHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	// Bootstrap check: allow unauthenticated creation ONLY if no tokens exist at all
	var count int
	err := h.db.DB.QueryRowContext(r.Context(), "SELECT COUNT(id) FROM tokens").Scan(&count)
	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "database_error", "Failed to check token count")
		return
	}

	actorID := "bootstrap"
	if count > 0 {
		// Requires valid token in context from middleware
		record := ctxutil.GetTokenRecord(r.Context())
		if record == nil {
			response.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Token required to create new tokens")
			return
		}
		
		// Enforce admin scope
		isAdmin := false
		for _, p := range record.Permissions {
			if p == "admin" {
				isAdmin = true
				break
			}
		}
		if !isAdmin {
			response.WriteError(w, r, http.StatusForbidden, "forbidden", "Admin permission required to create tokens")
			return
		}
		actorID = record.ID
	}

	var req CreateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if len(req.Permissions) == 0 {
		req.Permissions = []string{"admin"} // Default for bootstrap
	}

	plainToken, record, err := h.authService.CreateToken(r.Context(), req.Description, req.Permissions, nil)
	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "internal_error", "Failed to create token")
		return
	}

	h.auditService.Log(r.Context(), audit.Event{
		TokenID:    actorID,
		Action:     "token.create",
		Target:     record.ID,
		Parameters: map[string]interface{}{"permissions": req.Permissions, "description": req.Description},
		Result:     "success",
		ClientIP:   r.RemoteAddr,
	})

	response.WriteJSON(w, http.StatusCreated, CreateTokenResponse{
		Token: plainToken, // Never shown again
		ID:    record.ID,
		Perms: record.Permissions,
	})
}
