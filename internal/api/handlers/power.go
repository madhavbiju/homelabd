package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/madhavbiju/homelabd/internal/api/response"
	"github.com/madhavbiju/homelabd/internal/audit"
	"github.com/madhavbiju/homelabd/internal/ctxutil"
	"github.com/madhavbiju/homelabd/internal/power"
)

type PowerHandler struct {
	svc      *power.Service
	auditSvc *audit.Service
}

func NewPowerHandler(svc *power.Service, auditSvc *audit.Service) *PowerHandler {
	return &PowerHandler{svc: svc, auditSvc: auditSvc}
}

type PowerRequest struct {
	Confirm bool `json:"confirm"`
}

func (h *PowerHandler) Operate(w http.ResponseWriter, r *http.Request) {
	// The path determines the action: /api/v1/system/reboot or /shutdown
	pathParts := strings.Split(r.URL.Path, "/")
	action := pathParts[len(pathParts)-1] // 'reboot' or 'shutdown'

	if action != "reboot" && action != "shutdown" {
		response.WriteError(w, r, http.StatusNotFound, "not_found", "Action not found")
		return
	}

	var req PowerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if !req.Confirm {
		response.WriteError(w, r, http.StatusPreconditionFailed, "confirmation_required", "The 'confirm: true' parameter is required to execute power operations")
		return
	}

	actorID := "unknown"
	if record := ctxutil.GetTokenRecord(r.Context()); record != nil {
		actorID = record.ID
	}

	var err error
	if action == "reboot" {
		err = h.svc.Reboot(r.Context())
	} else {
		err = h.svc.Shutdown(r.Context())
	}

	result := "success"
	errMsg := ""
	if err != nil {
		result = "failure"
		errMsg = err.Error()
	}

	// Audit Log
	h.auditSvc.Log(r.Context(), audit.Event{
		TokenID:      actorID,
		Action:       "system." + action,
		Target:       "host",
		Result:       result,
		ErrorMessage: errMsg,
		ClientIP:     r.RemoteAddr,
	})

	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "power_error", "Failed to execute power operation")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "success", "action": action, "message": "Power operation scheduled successfully"})
}
