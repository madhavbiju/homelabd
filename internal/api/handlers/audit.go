package handlers

import (
	"net/http"
	"strconv"

	"github.com/madhavbiju/homelabd/internal/api/response"
	"github.com/madhavbiju/homelabd/internal/audit"
)

type AuditHandler struct {
	svc *audit.Service
}

func NewAuditHandler(svc *audit.Service) *AuditHandler {
	return &AuditHandler{svc: svc}
}

func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0

	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}
	if o, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && o >= 0 {
		offset = o
	}

	logs, err := h.svc.List(r.Context(), limit, offset)
	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "audit_error", "Failed to list audit logs")
		return
	}

	response.WriteJSON(w, http.StatusOK, logs)
}
