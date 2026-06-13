package responses

type AtsEvaluationResponse struct {
	ID        string  `json:"id"`
	ResumeID  string  `json:"resumeId"`
	JobID     string  `json:"jobId"`
	Score     float64 `json:"score"`
	Summary   string  `json:"summary"`
	Details   string  `json:"details"`
	CreatedAt string  `json:"createdAt"`
}

type AtsEvaluationSummaryResponse struct {
	ID        string  `json:"id"`
	ResumeID  string  `json:"resumeId"`
	JobID     string  `json:"jobId"`
	Score     float64 `json:"score"`
	Summary   string  `json:"summary"`
	CreatedAt string  `json:"createdAt"`
}
