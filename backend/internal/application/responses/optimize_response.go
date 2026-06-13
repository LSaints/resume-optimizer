package responses

type OptimizeResponse struct {
	ID           string `json:"id"`
	ResumeID     string `json:"resumeId"`
	JobID        string `json:"jobId"`
	TypstContent string `json:"typstContent"`
	CreatedAt    string `json:"createdAt"`
}

type OptimizeSummaryResponse struct {
	ID        string `json:"id"`
	ResumeID  string `json:"resumeId"`
	JobID     string `json:"jobId"`
	CreatedAt string `json:"createdAt"`
}
