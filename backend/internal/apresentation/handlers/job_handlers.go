package handlers

import (
	"encoding/json"
	"net/http"

	"backend/internal/application/requests"
	"backend/internal/application/services"
	"backend/internal/apresentation/middleware"
)

type JobHandler struct {
	Service *services.JobServices
}

func NewJobHandler(service *services.JobServices) *JobHandler {
	return &JobHandler{Service: service}
}

func (h *JobHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Context().Value(middleware.UserIDKey).(string)

	var req requests.CreateJobRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "json inválido", http.StatusBadRequest)
		return
	}

	job, err := h.Service.Create(userID, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "título é obrigatório" || err.Error() == "descrição é obrigatória" {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)
}

func (h *JobHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Context().Value(middleware.UserIDKey).(string)

	jobs, err := h.Service.GetByUserID(userID)
	if err != nil {
		http.Error(w, "erro ao listar vagas", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jobs)
}

func (h *JobHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Context().Value(middleware.UserIDKey).(string)
	id := r.PathValue("id")

	job, err := h.Service.GetByID(userID, id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "vaga não encontrada" {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(job)
}

func (h *JobHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Context().Value(middleware.UserIDKey).(string)
	id := r.PathValue("id")

	var req requests.UpdateJobRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "json inválido", http.StatusBadRequest)
		return
	}

	job, err := h.Service.Update(userID, id, req)
	if err != nil {
		status := http.StatusInternalServerError
		msg := err.Error()
		if msg == "vaga não encontrada" {
			status = http.StatusNotFound
		} else if msg == "título é obrigatório" || msg == "descrição é obrigatória" {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(job)
}

func (h *JobHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Context().Value(middleware.UserIDKey).(string)
	id := r.PathValue("id")

	err := h.Service.Delete(userID, id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "vaga não encontrada" {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
