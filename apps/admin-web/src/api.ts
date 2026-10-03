export type OrganizationKind = 'enterprise' | 'department'

export type Member = {
  id: string
  userId: string
  organizationId: string
  loginId: string
  status: 'pending' | 'active' | 'suspended' | 'revoked'
  roles: string[]
  organizationKind?: OrganizationKind
  active?: boolean
}

export type SessionResponse = { member: Member }

async function request<T>(path: string, init: RequestInit = {}, retry = true): Promise<T> {
  const response = await fetch(path, {
    ...init,
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...(init.headers ?? {}) },
  })
  if (response.status === 401 && retry && !path.includes('/auth/')) {
    await request('/v1/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ client: 'web' }),
    }, false)
    return request<T>(path, init, false)
  }
  if (!response.ok) {
    const error = new Error(`HTTP ${response.status}`)
    ;(error as Error & { status?: number }).status = response.status
    throw error
  }
  return response.status === 204 ? (undefined as T) : response.json() as Promise<T>
}

export const api = {
  login(loginId: string, password: string) {
    return request<SessionResponse>('/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ loginId, password, client: 'web' }),
    })
  },
  me() {
    return request<Member>('/v1/me')
  },
  logout() {
    return request<void>('/v1/auth/logout', { method: 'POST' })
  },
  refresh() {
    return request<SessionResponse>('/v1/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ client: 'web' }),
    })
  },
}
