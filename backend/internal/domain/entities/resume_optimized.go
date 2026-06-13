package entities

import (
	"time"

	"github.com/google/uuid"
)

type ResumeOptimized struct {
	ID        uuid.UUID `json:"id"`
	ResumeID  uuid.UUID `json:"resumeId"`
	RawText   string    `json:"rawText"`
	CreatedAt time.Time `json:"createdAt"`
}
