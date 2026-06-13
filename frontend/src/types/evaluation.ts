export interface EvaluateRequest {
  jobId: string
}

export interface EvaluationResponse {
  id: string
  resumeId: string
  jobId: string
  score: number
  summary: string
  details: string
  createdAt: string
}

export interface EvaluationSummaryResponse {
  id: string
  resumeId: string
  jobId: string
  score: number
  summary: string
  createdAt: string
}
