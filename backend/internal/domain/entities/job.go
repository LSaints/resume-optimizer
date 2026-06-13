package entities

import (
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	RawDescription string    `json:"rawDescription"`
	CreatedAt      time.Time `json:"createdAt"`
}
