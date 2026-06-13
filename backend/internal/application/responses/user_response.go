package responses

import "github.com/google/uuid"

type UserResponse struct {
	Id    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}
