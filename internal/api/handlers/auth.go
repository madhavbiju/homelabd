package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/madhavbiju/homelabd/internal/api/response"
	"github.com/madhavbiju/homelabd/internal/ctxutil"
	"github.com/madhavbiju/homelabd/internal/audit"
	"github.com/madhavbiju/homelabd/internal/auth"
	"github.com/madhavbiju/homelabd/internal/database"
	"github.com/madhavbiju/homelabd/internal/users"
)

type AuthHandler struct {
	authService  *auth.Service
	auditService *audit.Service
	usersService *users.Service
	db           *database.Database
}

func NewAuthHandler(authService *auth.Service, auditService *audit.Service, usersService *users.Service, db *database.Database) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		auditService: auditService,
		usersService: usersService,
		db:           db,
	}
}

type SetupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Setup(w http.ResponseWriter, r *http.Request) {
	count, err := h.usersService.Count(r.Context())
	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "database_error", "Failed to check user count")
		return
	}

	if count > 0 {
		response.WriteError(w, r, http.StatusForbidden, "forbidden", "Setup already completed")
		return
	}

	var req SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if req.Username == "" || req.Password == "" {
		response.WriteError(w, r, http.StatusBadRequest, "invalid_request", "Username and password required")
		return
	}

	user, err := h.usersService.CreateUser(r.Context(), req.Username, req.Password, "admin")
	if err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "setup_error", err.Error())
		return
	}

	// Create an initial session token
	plainToken, record, err := h.authService.CreateToken(r.Context(), &user.ID, "Initial Admin Session", []string{"admin"}, nil)
	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "internal_error", "Failed to create token")
		return
	}

	h.auditService.Log(r.Context(), audit.Event{
		TokenID:  record.ID,
		Action:   "system.setup",
		Target:   user.ID,
		Result:   "success",
		ClientIP: r.RemoteAddr,
	})

	response.WriteJSON(w, http.StatusCreated, map[string]string{
		"message": "Setup complete",
		"token":   plainToken,
	})
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	user, err := h.usersService.VerifyPassword(r.Context(), req.Username, req.Password)
	if err != nil {
		response.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Invalid username or password")
		return
	}

	// Create a session token (in a real app, maybe expire after 24h)
	plainToken, _, err := h.authService.CreateToken(r.Context(), &user.ID, "Session token", []string{user.Role}, nil)
	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "internal_error", "Failed to create session token")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{
		"token": plainToken,
	})
}

type CreateTokenRequest struct {
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	UserID      *string  `json:"user_id"` // Optional
}

func (h *AuthHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
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

	var req CreateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if len(req.Permissions) == 0 {
		req.Permissions = []string{"admin"}
	}

	plainToken, newRecord, err := h.authService.CreateToken(r.Context(), req.UserID, req.Description, req.Permissions, nil)
	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "internal_error", "Failed to create token")
		return
	}

	h.auditService.Log(r.Context(), audit.Event{
		TokenID:    record.ID,
		Action:     "token.create",
		Target:     newRecord.ID,
		Parameters: map[string]interface{}{"permissions": req.Permissions, "description": req.Description},
		Result:     "success",
		ClientIP:   r.RemoteAddr,
	})

	response.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"token":       plainToken,
		"id":          newRecord.ID,
		"permissions": newRecord.Permissions,
	})
}
