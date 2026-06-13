function decodeTokenPayload(token: string): Record<string, unknown> | null {
  try {
    const parts = token.split('.')
    if (parts.length !== 3) return null
    const payload = parts[1]
    const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'))
    return JSON.parse(decoded)
  } catch {
    return null
  }
}

export function extractUserId(token: string): string | null {
  const payload = decodeTokenPayload(token)
  if (!payload) return null
  const userId = payload['userID']
  return typeof userId === 'string' ? userId : null
}

export function isTokenExpired(token: string): boolean {
  const payload = decodeTokenPayload(token)
  if (!payload) return true
  const exp = payload['exp']
  if (typeof exp !== 'number') return true
  return Date.now() >= exp * 1000
}
