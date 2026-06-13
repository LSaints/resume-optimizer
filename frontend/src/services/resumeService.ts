import type { ResumeResponse } from '../types/resume'
import { get, post, del } from './api'

export function list(): Promise<ResumeResponse[]> {
  return get<ResumeResponse[]>('/resumes')
}

export function getById(id: string): Promise<ResumeResponse> {
  return get<ResumeResponse>(`/resumes/${id}`)
}

export function upload(file: File): Promise<ResumeResponse> {
  const formData = new FormData()
  formData.append('file', file)
  return post<ResumeResponse>('/resumes', formData, true)
}

export function remove(id: string): Promise<void> {
  return del<void>(`/resumes/${id}`)
}
