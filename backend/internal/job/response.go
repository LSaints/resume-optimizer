package job

import (
	"time"

	"github.com/google/uuid"
)

type JobResponse struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	RawDescription string    `json:"rawDescription"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
