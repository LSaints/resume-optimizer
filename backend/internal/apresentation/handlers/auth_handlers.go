package handlers

import (
	"encoding/json"
	"net/http"

	"backend/internal/application/requests"
	"backend/internal/application/services"
)

type AuthHandler struct {
	Service *services.AuthServices
}

func NewAuthHandler(
	service *services.AuthServices,
) *AuthHandler {
	return &AuthHandler{
		Service: service,
	}
}

func (h *AuthHandler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	var request requests.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(
			w,
			"json inválido",
			http.StatusBadRequest,
		)
		return
	}

	response, err := h.Service.Login(request.Email, request.Password)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusUnauthorized,
		)
		return
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(
			w,
			"erro ao serializar resposta",
			http.StatusInternalServerError,
		)
		return
	}
}
