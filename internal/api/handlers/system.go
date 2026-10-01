package handlers

import (
	"net/http"
	"strconv"

	"github.com/madhavbiju/homelabd/internal/api/response"
	"github.com/madhavbiju/homelabd/internal/system"
)

type SystemHandler struct {
	svc *system.Service
}

func NewSystemHandler(svc *system.Service) *SystemHandler {
	return &SystemHandler{svc: svc}
}

func (h *SystemHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	info, err := h.svc.GetInfo(r.Context())
	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "system_error", "Failed to get system info")
		return
	}
	response.WriteJSON(w, http.StatusOK, info)
}

func (h *SystemHandler) GetResources(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.GetResources(r.Context())
	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "system_error", "Failed to get system resources")
		return
	}
	response.WriteJSON(w, http.StatusOK, res)
}

func (h *SystemHandler) GetStorage(w http.ResponseWriter, r *http.Request) {
	storage, err := h.svc.GetStorage(r.Context())
	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "system_error", "Failed to get storage info")
		return
	}
	response.WriteJSON(w, http.StatusOK, storage)
}

func (h *SystemHandler) GetNetwork(w http.ResponseWriter, r *http.Request) {
	netInfo, err := h.svc.GetNetwork(r.Context())
	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "system_error", "Failed to get network info")
		return
	}
	response.WriteJSON(w, http.StatusOK, netInfo)
}

func (h *SystemHandler) GetProcesses(w http.ResponseWriter, r *http.Request) {
	limit := 100 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	procs, err := h.svc.GetProcesses(r.Context(), limit)
	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "system_error", "Failed to get processes")
		return
	}
	response.WriteJSON(w, http.StatusOK, procs)
}
