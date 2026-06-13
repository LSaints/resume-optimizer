import type { OptimizeRequest, OptimizeResponse, OptimizeSummaryResponse } from '../types/optimization'
import { get, post, del } from './api'

export function optimize(resumeId: string, jobId: string): Promise<OptimizeResponse> {
  const body: OptimizeRequest = { jobId }
  return post<OptimizeResponse>(`/resumes/${resumeId}/optimize`, body)
}

export function listByResume(resumeId: string): Promise<OptimizeSummaryResponse[]> {
  return get<OptimizeSummaryResponse[]>(`/resumes/${resumeId}/optimizations`)
}

export function getByID(resumeId: string, optimizationId: string): Promise<OptimizeResponse> {
  return get<OptimizeResponse>(`/resumes/${resumeId}/optimizations/${optimizationId}`)
}

export function remove(resumeId: string, optimizationId: string): Promise<void> {
  return del<void>(`/resumes/${resumeId}/optimizations/${optimizationId}`)
}
