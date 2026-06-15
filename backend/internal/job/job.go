package job

import (
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"userId"`
	Title          string    `json:"title"`
	RawDescription string    `json:"rawDescription"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
