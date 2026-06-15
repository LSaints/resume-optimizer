package user

import (
	"encoding/json"
	"net/http"
)

type UserHandler struct {
	Service *UserServices
}

func NewUserHandler(
	service *UserServices,
) *UserHandler {
	return &UserHandler{
		Service: service,
	}
}

func (h *UserHandler) GetUsers(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	users, err := h.Service.GetUsers()
	if err != nil {
		http.Error(w, "Erro ao buscar usuarios", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(
			w,
			"erro ao serializar resposta",
			http.StatusInternalServerError,
		)
		return
	}
}

func (h *UserHandler) GetUsersById(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	id := r.PathValue("id")

	user, err := h.Service.GetUserById(id)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusNotFound,
		)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) CreateUser(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	var request CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(
			w,
			"json inválido",
			http.StatusBadRequest,
		)
		return
	}

	user, err := h.Service.CreateUser(request)
	if err != nil {
		http.Error(w, "Erro ao buscar usuarios", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(
			w,
			"erro ao serializar resposta",
			http.StatusInternalServerError,
		)
		return
	}
}
