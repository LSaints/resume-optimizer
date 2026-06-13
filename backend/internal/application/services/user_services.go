package services

import (
	"backend/internal/application/requests"
	"backend/internal/application/responses"
	"backend/internal/domain/entities"
	"backend/internal/infrastructure/repositories"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserServices struct {
	Repository *repositories.UserRepository
}

func NewUserServices(
	repository *repositories.UserRepository,
) *UserServices {
	return &UserServices{
		Repository: repository,
	}
}

func (s *UserServices) GetUsers() ([]responses.UserResponse, error) {
	users, err := s.Repository.GetUsers()
	if err != nil {
		return nil, err
	}

	result := make([]responses.UserResponse, 0, len(users))

	for _, user := range users {
		result = append(result, responses.UserResponse{
			Id:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		})
	}

	return result, nil
}

func (s *UserServices) GetUserById(
	id string,
) (responses.UserResponse, error) {

	user, err := s.Repository.GetUserById(id)
	if err != nil {
		return responses.UserResponse{}, err
	}

	return responses.UserResponse{
		Id:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s *UserServices) CreateUser(
	request requests.CreateUserRequest,
) (responses.UserResponse, error) {

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(request.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return responses.UserResponse{}, err
	}

	user := entities.User{
		ID:       uuid.New(),
		Name:     request.Name,
		Email:    request.Email,
		Password: string(hash),
	}

	err = s.Repository.CreateUser(user)
	if err != nil {
		return responses.UserResponse{}, err
	}

	return responses.UserResponse{
		Id:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}
