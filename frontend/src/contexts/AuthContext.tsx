import { createContext, useState, useEffect, useCallback, type ReactNode } from 'react'
import type { UserResponse } from '../types/user'
import { getToken, setToken, removeToken } from '../utils/storage'
import { extractUserId, isTokenExpired } from '../utils/jwt'
import * as authService from '../services/authService'

export interface AuthContextValue {
  user: UserResponse | null
  token: string | null
  isAuthenticated: boolean
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (name: string, email: string, password: string) => Promise<void>
  logout: () => void
}

export const AuthContext = createContext<AuthContextValue | undefined>(undefined)

interface AuthProviderProps {
  children: ReactNode
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<UserResponse | null>(null)
  const [token, setTokenState] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const savedToken = getToken()
    if (!savedToken || isTokenExpired(savedToken)) {
      removeToken()
      setLoading(false)
      return
    }

    const userId = extractUserId(savedToken)
    if (!userId) {
      removeToken()
      setLoading(false)
      return
    }

    setTokenState(savedToken)

    authService
      .getMe(userId)
      .then((userData) => {
        setUser(userData)
      })
      .catch(() => {
        removeToken()
        setTokenState(null)
      })
      .finally(() => {
        setLoading(false)
      })
  }, [])

  const login = useCallback(async (email: string, password: string) => {
    const { token: newToken } = await authService.login({ email, password })
    setToken(newToken)
    setTokenState(newToken)

    const userId = extractUserId(newToken)
    if (!userId) throw new Error('Token inválido')

    const userData = await authService.getMe(userId)
    setUser(userData)
  }, [])

  const register = useCallback(async (name: string, email: string, password: string) => {
    await authService.register({ name, email, password })
    await login(email, password)
  }, [login])

  const logout = useCallback(() => {
    removeToken()
    setTokenState(null)
    setUser(null)
  }, [])

  const isAuthenticated = !!token && !!user

  return (
    <AuthContext.Provider value={{ user, token, isAuthenticated, loading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  )
}
