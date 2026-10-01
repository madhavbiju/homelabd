package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/madhavbiju/homelabd/internal/api/response"
	"github.com/madhavbiju/homelabd/internal/audit"
	"github.com/madhavbiju/homelabd/internal/compose"
	"github.com/madhavbiju/homelabd/internal/ctxutil"
)

type ComposeHandler struct {
	svc      *compose.Service
	auditSvc *audit.Service
}

func NewComposeHandler(svc *compose.Service, auditSvc *audit.Service) *ComposeHandler {
	return &ComposeHandler{svc: svc, auditSvc: auditSvc}
}

func (h *ComposeHandler) ListStacks(w http.ResponseWriter, r *http.Request) {
	stacks, err := h.svc.ListStacks(r.Context())
	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "compose_error", "Failed to list stacks")
		return
	}
	response.WriteJSON(w, http.StatusOK, stacks)
}

func (h *ComposeHandler) GetStack(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	stack, err := h.svc.GetStack(r.Context(), name)
	if err != nil {
		response.WriteError(w, r, http.StatusNotFound, "not_found", "Stack not found")
		return
	}
	response.WriteJSON(w, http.StatusOK, stack)
}

func (h *ComposeHandler) OperateStack(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	op := chi.URLParam(r, "operation") // up, down, pull

	if op != "up" && op != "down" && op != "pull" {
		response.WriteError(w, r, http.StatusBadRequest, "invalid_operation", "Operation must be up, down, or pull")
		return
	}

	actorID := "unknown"
	if record := ctxutil.GetTokenRecord(r.Context()); record != nil {
		actorID = record.ID
	}

	var err error
	switch op {
	case "up":
		err = h.svc.Up(r.Context(), name)
	case "down":
		err = h.svc.Down(r.Context(), name)
	case "pull":
		err = h.svc.Pull(r.Context(), name)
	}

	result := "success"
	errMsg := ""
	if err != nil {
		result = "failure"
		errMsg = err.Error()
	}

	// Audit log
	h.auditSvc.Log(r.Context(), audit.Event{
		TokenID:      actorID,
		Action:       "compose." + op,
		Target:       name,
		Result:       result,
		ErrorMessage: errMsg,
		ClientIP:     r.RemoteAddr,
	})

	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "compose_error", "Failed to perform operation: "+err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "success", "operation": op})
}
