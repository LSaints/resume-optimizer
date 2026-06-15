package resumeoptimized

import "time"

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

func (s *OptimizationServices) toResponse(opt ResumeOptimized) OptimizeResponse {
	return OptimizeResponse{
		ID:           opt.ID.String(),
		ResumeID:     opt.ResumeID.String(),
		JobID:        opt.JobID.String(),
		TypstContent: opt.TypstContent,
		CreatedAt:    opt.CreatedAt.Format(time.RFC3339),
	}
}
