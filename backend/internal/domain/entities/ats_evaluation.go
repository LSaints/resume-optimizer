package entities

import (
	"time"

	"github.com/google/uuid"
)

type AtsEvaluation struct {
	ID          uuid.UUID `json:"id"`
	ResumeID    uuid.UUID `json:"resumeId"`
	JobID       uuid.UUID `json:"jobId"`
	Score       float64   `json:"score"`
	Summary     string    `json:"summary"`
	Details     string    `json:"details"`
	RawResponse string    `json:"rawResponse"`
	CreatedAt   time.Time `json:"createdAt"`
}
