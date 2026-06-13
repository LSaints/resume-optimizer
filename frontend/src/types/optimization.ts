export interface OptimizeRequest {
  jobId: string
}

export interface OptimizeResponse {
  id: string
  resumeId: string
  jobId: string
  typstContent: string
  createdAt: string
}

export interface OptimizeSummaryResponse {
  id: string
  resumeId: string
  jobId: string
  createdAt: string
}
