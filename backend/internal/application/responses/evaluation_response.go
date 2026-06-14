package responses

type AtsEvaluationResponse struct {
	ID                    string   `json:"id"`
	ResumeID              string   `json:"resumeId"`
	JobID                 string   `json:"jobId"`
	Score                 float64  `json:"score"`
	Summary               string   `json:"summary"`
	Details               string   `json:"details"`
	BreakdownKeywordMatch float64  `json:"breakdownKeywordMatch"`
	BreakdownTechnical    float64  `json:"breakdownTechnical"`
	BreakdownExperience   float64  `json:"breakdownExperience"`
	BreakdownImpact       float64  `json:"breakdownImpact"`
	BreakdownReadability  float64  `json:"breakdownReadability"`
	MatchedKeywords       []string `json:"matchedKeywords"`
	MissingKeywords       []string `json:"missingKeywords"`
	Recommendations       []string `json:"recommendations"`
	CreatedAt             string   `json:"createdAt"`
}

type AtsEvaluationSummaryResponse struct {
	ID        string  `json:"id"`
	ResumeID  string  `json:"resumeId"`
	JobID     string  `json:"jobId"`
	Score     float64 `json:"score"`
	Summary   string  `json:"summary"`
	CreatedAt string  `json:"createdAt"`
}
