package handlers

import (
	"encoding/json"
	"net/http"

	"backend/internal/application/requests"
	"backend/internal/application/services"
	"backend/internal/apresentation/middleware"
)

type AtsScoringHandler struct {
	Service *services.AtsScoringServices
}

func NewAtsScoringHandler(service *services.AtsScoringServices) *AtsScoringHandler {
	return &AtsScoringHandler{Service: service}
}

func (h *AtsScoringHandler) Evaluate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "não autorizado", http.StatusUnauthorized)
		return
	}
	resumeID := r.PathValue("resumeID")

	var req requests.EvaluateResumeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "json inválido", http.StatusBadRequest)
		return
	}

	if req.JobID == "" {
		http.Error(w, "jobId é obrigatório", http.StatusBadRequest)
		return
	}

	result, err := h.Service.Evaluate(userID, resumeID, req.JobID)
	if err != nil {
		status := http.StatusInternalServerError
		msg := err.Error()
		switch msg {
		case "currículo não encontrado", "vaga não encontrada":
			status = http.StatusNotFound
		case "serviço de IA não configurado":
			status = http.StatusInternalServerError
		case "erro ao processar avaliação":
			status = http.StatusBadGateway
		case "erro ao comunicar com a API", "erro ao processar otimização":
			status = http.StatusBadGateway
		}
		http.Error(w, msg, status)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (h *AtsScoringHandler) ListByResume(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "não autorizado", http.StatusUnauthorized)
		return
	}
	resumeID := r.PathValue("resumeID")

	results, err := h.Service.ListByResume(userID, resumeID)
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

func (h *AtsScoringHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "não autorizado", http.StatusUnauthorized)
		return
	}
	evaluationID := r.PathValue("evaluationID")

	result, err := h.Service.GetByID(userID, evaluationID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "avaliação não encontrada" {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
