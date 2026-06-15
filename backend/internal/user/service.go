package user

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserServices struct {
	Repository *UserRepository
}

func NewUserServices(
	repository *UserRepository,
) *UserServices {
	return &UserServices{
		Repository: repository,
	}
}

func (s *UserServices) GetUsers() ([]UserResponse, error) {
	users, err := s.Repository.GetUsers()
	if err != nil {
		return nil, err
	}

	result := make([]UserResponse, 0, len(users))

	for _, user := range users {
		result = append(result, UserResponse{
			Id:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		})
	}

	return result, nil
}

func (s *UserServices) GetUserById(
	id string,
) (UserResponse, error) {

	user, err := s.Repository.GetUserById(id)
	if err != nil {
		return UserResponse{}, err
	}

	return UserResponse{
		Id:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s *UserServices) CreateUser(
	request CreateUserRequest,
) (UserResponse, error) {

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(request.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return UserResponse{}, err
	}

	user := User{
		ID:       uuid.New(),
		Name:     request.Name,
		Email:    request.Email,
		Password: string(hash),
	}

	err = s.Repository.CreateUser(user)
	if err != nil {
		return UserResponse{}, err
	}

	return UserResponse{
		Id:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}
