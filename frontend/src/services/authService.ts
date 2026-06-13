import type { LoginRequest, RegisterRequest, LoginResponse } from '../types/auth'
import type { UserResponse } from '../types/user'
import { post, get } from './api'

export async function login(data: LoginRequest): Promise<LoginResponse> {
  return post<LoginResponse>('/auth/login', data)
}

export async function register(data: RegisterRequest): Promise<UserResponse> {
  return post<UserResponse>('/users', data)
}

export async function getMe(userId: string): Promise<UserResponse> {
  return get<UserResponse>(`/users/${userId}`)
}
