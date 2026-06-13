export interface JobRequest {
  title: string
  rawDescription: string
}

export interface JobResponse {
  id: string
  userId: string
  title: string
  rawDescription: string
  createdAt: string
  updatedAt: string
}
