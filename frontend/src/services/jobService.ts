import type { JobRequest, JobResponse } from '../types/job'
import { get, post, put, del } from './api'

export function list(): Promise<JobResponse[]> {
  return get<JobResponse[]>('/jobs')
}

export function getById(id: string): Promise<JobResponse> {
  return get<JobResponse>(`/jobs/${id}`)
}

export function create(data: JobRequest): Promise<JobResponse> {
  return post<JobResponse>('/jobs', data)
}

export function update(id: string, data: JobRequest): Promise<JobResponse> {
  return put<JobResponse>(`/jobs/${id}`, data)
}

export function remove(id: string): Promise<void> {
  return del<void>(`/jobs/${id}`)
}
