package entities

import (
	"time"

	"github.com/google/uuid"
)

type ResumeOptimized struct {
	ID           uuid.UUID `json:"id"`
	ResumeID     uuid.UUID `json:"resumeId"`
	JobID        uuid.UUID `json:"jobId"`
	SystemPrompt string    `json:"systemPrompt"`
	UserPrompt   string    `json:"userPrompt"`
	RawText      string    `json:"rawText"`
	TypstContent string    `json:"typstContent"`
	CreatedAt    time.Time `json:"createdAt"`
}
