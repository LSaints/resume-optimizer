import type { EvaluateRequest, EvaluationResponse, EvaluationSummaryResponse } from '../types/evaluation'
import { get, post } from './api'

export function evaluate(resumeId: string, jobId: string): Promise<EvaluationResponse> {
  const body: EvaluateRequest = { jobId }
  return post<EvaluationResponse>(`/resumes/${resumeId}/evaluate`, body)
}

export function listByResume(resumeId: string): Promise<EvaluationSummaryResponse[]> {
  return get<EvaluationSummaryResponse[]>(`/resumes/${resumeId}/evaluations`)
}

export function getByID(resumeId: string, evaluationId: string): Promise<EvaluationResponse> {
  return get<EvaluationResponse>(`/resumes/${resumeId}/evaluations/${evaluationId}`)
}
