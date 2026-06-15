package resume

import (
	"encoding/json"
	"net/http"
	"strings"

	"backend/internal/auth"
)

const maxUploadSize = 10 << 20

type ResumeHandler struct {
	Service *ResumeServices
}

func NewResumeHandler(service *ResumeServices) *ResumeHandler {
	return &ResumeHandler{
		Service: service,
	}
}

func (h *ResumeHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "não autorizado", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "arquivo muito grande", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "arquivo não enviado", http.StatusBadRequest)
		return
	}
	defer file.Close()

	originalName := header.Filename

	if !isValidExtension(originalName) {
		http.Error(w, "formato de arquivo não suportado", http.StatusBadRequest)
		return
	}

	resume, err := h.Service.Create(userID, originalName, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resume)
}

func (h *ResumeHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "não autorizado", http.StatusUnauthorized)
		return
	}

	resumes, err := h.Service.GetByUserID(userID)
	if err != nil {
		http.Error(w, "erro ao listar currículos", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resumes)
}

func (h *ResumeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "não autorizado", http.StatusUnauthorized)
		return
	}
	id := r.PathValue("id")

	resume, err := h.Service.GetByID(userID, id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "currículo não encontrado" {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resume)
}

func (h *ResumeHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "não autorizado", http.StatusUnauthorized)
		return
	}
	id := r.PathValue("id")

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "arquivo muito grande", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "arquivo não enviado", http.StatusBadRequest)
		return
	}
	defer file.Close()

	originalName := header.Filename

	if !isValidExtension(originalName) {
		http.Error(w, "formato de arquivo não suportado", http.StatusBadRequest)
		return
	}

	resume, err := h.Service.Update(userID, id, originalName, file)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "currículo não encontrado" {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resume)
}

func (h *ResumeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "não autorizado", http.StatusUnauthorized)
		return
	}
	id := r.PathValue("id")

	err := h.Service.Delete(userID, id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "currículo não encontrado" {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func isValidExtension(filename string) bool {
	name := strings.ToLower(filename)
	return strings.HasSuffix(name, ".pdf") || strings.HasSuffix(name, ".docx")
}
