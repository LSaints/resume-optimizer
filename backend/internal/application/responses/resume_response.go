package responses

import (
	"time"

	"github.com/google/uuid"
)

type ResumeResponse struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"userId"`
	OriginalName string    `json:"originalName"`
	RawText      string    `json:"rawText"`
	UploadedAt   time.Time `json:"uploadedAt"`
}

type ResumeSummaryResponse struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"userId"`
	OriginalName string    `json:"originalName"`
	UploadedAt   time.Time `json:"uploadedAt"`
}
