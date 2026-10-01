package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/madhavbiju/homelabd/internal/api/response"
	"github.com/madhavbiju/homelabd/internal/audit"
	"github.com/madhavbiju/homelabd/internal/ctxutil"
	"github.com/madhavbiju/homelabd/internal/docker"
)

type DockerHandler struct {
	svc      *docker.Service
	auditSvc *audit.Service
}

func NewDockerHandler(svc *docker.Service, auditSvc *audit.Service) *DockerHandler {
	return &DockerHandler{svc: svc, auditSvc: auditSvc}
}

func (h *DockerHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	info, err := h.svc.Info(r.Context())
	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "docker_error", "Failed to get docker info")
		return
	}
	response.WriteJSON(w, http.StatusOK, info)
}

func (h *DockerHandler) ListContainers(w http.ResponseWriter, r *http.Request) {
	containers, err := h.svc.ListContainers(r.Context())
	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "docker_error", "Failed to list containers")
		return
	}
	response.WriteJSON(w, http.StatusOK, containers)
}

func (h *DockerHandler) GetContainer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	container, err := h.svc.GetContainer(r.Context(), id)
	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "docker_error", "Failed to get container")
		return
	}
	response.WriteJSON(w, http.StatusOK, container)
}

func (h *DockerHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tail := r.URL.Query().Get("tail")
	if tail == "" {
		tail = "100" // Set a safe bound by default
	}

	reader, err := h.svc.GetLogs(r.Context(), id, tail)
	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "docker_error", "Failed to get logs")
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", "text/plain")
	io.Copy(w, reader)
}

func (h *DockerHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	reader, err := h.svc.GetStats(r.Context(), id)
	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "docker_error", "Failed to get stats")
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, reader)
}

func (h *DockerHandler) OperateContainer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	op := chi.URLParam(r, "operation") // start, stop, restart

	if op != "start" && op != "stop" && op != "restart" {
		response.WriteError(w, r, http.StatusBadRequest, "invalid_operation", "Operation must be start, stop, or restart")
		return
	}

	actorID := "unknown"
	if record := ctxutil.GetTokenRecord(r.Context()); record != nil {
		actorID = record.ID
	}

	err := h.svc.OperateContainer(r.Context(), id, op)
	
	result := "success"
	errMsg := ""
	if err != nil {
		result = "failure"
		errMsg = err.Error()
	}

	// Audit log
	h.auditSvc.Log(r.Context(), audit.Event{
		TokenID:      actorID,
		Action:       "docker." + op,
		Target:       id,
		Result:       result,
		ErrorMessage: errMsg,
		ClientIP:     r.RemoteAddr,
	})

	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "docker_error", "Failed to perform operation")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "success", "operation": op})
}

func (h *DockerHandler) ListImages(w http.ResponseWriter, r *http.Request) {
	images, err := h.svc.ListImages(r.Context())
	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "docker_error", "Failed to list images")
		return
	}
	response.WriteJSON(w, http.StatusOK, images)
}

func (h *DockerHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	img, err := h.svc.GetImage(r.Context(), id)
	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "docker_error", "Failed to get image")
		return
	}
	response.WriteJSON(w, http.StatusOK, img)
}

func (h *DockerHandler) PullImage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Image string `json:"image"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Image == "" {
		response.WriteError(w, r, http.StatusBadRequest, "invalid_request", "Missing image name")
		return
	}

	actorID := "unknown"
	if record := ctxutil.GetTokenRecord(r.Context()); record != nil {
		actorID = record.ID
	}

	reader, err := h.svc.PullImage(r.Context(), req.Image)
	
	result := "success"
	errMsg := ""
	if err != nil {
		result = "failure"
		errMsg = err.Error()
	}

	h.auditSvc.Log(r.Context(), audit.Event{
		TokenID:      actorID,
		Action:       "docker.pull",
		Target:       req.Image,
		Result:       result,
		ErrorMessage: errMsg,
		ClientIP:     r.RemoteAddr,
	})

	if err != nil {
		response.WriteError(w, r, http.StatusInternalServerError, "docker_error", "Failed to pull image")
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, reader)
}
