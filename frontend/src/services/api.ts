import { getToken, removeToken } from '../utils/storage'

const BASE_URL = 'http://localhost:8080/v1'

interface ApiError {
  status: number
  message: string
}

function buildHeaders(isMultipart: boolean): HeadersInit {
  const headers: HeadersInit = {}
  if (!isMultipart) {
    headers['Content-Type'] = 'application/json'
  }
  const token = getToken()
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }
  return headers
}

function errorMessage(status: number): string {
  switch (status) {
    case 400:
      return 'Dados inválidos. Verifique os campos e tente novamente.'
    case 401:
      return 'Credenciais inválidas. Verifique seu email e senha.'
    case 403:
      return 'Você não tem permissão para acessar este recurso.'
    case 404:
      return 'Recurso não encontrado.'
    case 409:
      return 'Este email já está cadastrado.'
    case 413:
      return 'Arquivo muito grande. O limite é de 10MB.'
    case 429:
      return 'Muitas requisições. Aguarde um momento e tente novamente.'
    default:
      if (status >= 500) return 'Erro interno do servidor. Tente novamente mais tarde.'
      return 'Ocorreu um erro inesperado.'
  }
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  isMultipart = false,
): Promise<T> {
  const url = `${BASE_URL}${path}`
  const options: RequestInit = {
    method,
    headers: buildHeaders(isMultipart),
  }

  if (body !== undefined) {
    options.body = isMultipart ? body as FormData : JSON.stringify(body)
  }

  const response = await fetch(url, options)

  if (!response.ok) {
    if (response.status === 401) {
      removeToken()
    }
    throw { status: response.status, message: errorMessage(response.status) } as ApiError
  }

  if (response.status === 204) return undefined as T

  return response.json() as Promise<T>
}

export function get<T>(path: string): Promise<T> {
  return request<T>('GET', path)
}

export function post<T>(path: string, body?: unknown, isMultipart = false): Promise<T> {
  return request<T>('POST', path, body, isMultipart)
}

export function put<T>(path: string, body?: unknown): Promise<T> {
  return request<T>('PUT', path, body)
}

export function del<T = void>(path: string): Promise<T> {
  return request<T>('DELETE', path)
}
