package entities

import (
	"time"

	"github.com/google/uuid"
)

type AtsEvaluation struct {
	ID                    uuid.UUID `json:"id"`
	ResumeID              uuid.UUID `json:"resumeId"`
	JobID                 uuid.UUID `json:"jobId"`
	Score                 float64   `json:"score"`
	Summary               string    `json:"summary"`
	Details               string    `json:"details"`
	RawResponse           string    `json:"rawResponse"`
	BreakdownKeywordMatch float64   `json:"breakdownKeywordMatch"`
	BreakdownTechnical    float64   `json:"breakdownTechnical"`
	BreakdownExperience   float64   `json:"breakdownExperience"`
	BreakdownImpact       float64   `json:"breakdownImpact"`
	BreakdownReadability  float64   `json:"breakdownReadability"`
	MatchedKeywords       string    `json:"matchedKeywords"`
	MissingKeywords       string    `json:"missingKeywords"`
	Recommendations       string    `json:"recommendations"`
	CreatedAt             time.Time `json:"createdAt"`
}
