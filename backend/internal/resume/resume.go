package resume

import (
	"time"

	"github.com/google/uuid"
)

type Resume struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"userId"`
	OriginalName string    `json:"originalName"`
	RawText      string    `json:"rawText"`
	UploadedAt   time.Time `json:"uploadAt"`
}
