package handlers

import (
	"encoding/json"
	"net/http"

	"backend/internal/application/requests"
	"backend/internal/application/services"
	"backend/internal/apresentation/middleware"
)

type OptimizationHandler struct {
	Service *services.OptimizationServices
}

func NewOptimizationHandler(service *services.OptimizationServices) *OptimizationHandler {
	return &OptimizationHandler{Service: service}
}

func (h *OptimizationHandler) Optimize(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Context().Value(middleware.UserIDKey).(string)
	resumeID := r.PathValue("resumeID")

	var req requests.OptimizeResumeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "json inválido", http.StatusBadRequest)
		return
	}

	if req.JobID == "" {
		http.Error(w, "json inválido", http.StatusBadRequest)
		return
	}

	result, err := h.Service.Optimize(userID, resumeID, req.JobID)
	if err != nil {
		status := http.StatusInternalServerError
		msg := err.Error()
		switch msg {
		case "currículo não encontrado", "vaga não encontrada":
			status = http.StatusNotFound
		case "serviço de IA não configurado":
			status = http.StatusInternalServerError
		case "erro ao processar otimização":
			status = http.StatusBadGateway
		}
		http.Error(w, msg, status)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (h *OptimizationHandler) ListByResume(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Context().Value(middleware.UserIDKey).(string)
	resumeID := r.PathValue("resumeID")

	results, err := h.Service.GetByResumeID(userID, resumeID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "currículo não encontrado" {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}

func (h *OptimizationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Context().Value(middleware.UserIDKey).(string)
	optimizationID := r.PathValue("optimizationID")

	result, err := h.Service.GetByID(userID, optimizationID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "otimização não encontrada" {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
