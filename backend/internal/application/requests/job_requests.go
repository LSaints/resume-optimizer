package requests

type CreateJobRequest struct {
	Title          string `json:"title"`
	RawDescription string `json:"rawDescription"`
}

type UpdateJobRequest struct {
	Title          string `json:"title"`
	RawDescription string `json:"rawDescription"`
}
