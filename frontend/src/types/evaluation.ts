export interface EvaluateRequest {
  jobId: string
}

export interface AtsBreakdown {
  keywordMatch: number
  technicalCompatibility: number
  professionalExperience: number
  impactAndResults: number
  atsReadability: number
}

export interface EvaluationResponse {
  id: string
  resumeId: string
  jobId: string
  score: number
  summary: string
  details: string
  breakdownKeywordMatch: number
  breakdownTechnical: number
  breakdownExperience: number
  breakdownImpact: number
  breakdownReadability: number
  matchedKeywords: string[]
  missingKeywords: string[]
  recommendations: string[]
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
