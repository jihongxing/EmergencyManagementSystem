import { api, type Member, type OrganizationKind } from './api'

export type AuthState = 'loading' | 'unauthenticated' | 'authenticated' | 'forbidden'

export function workspaceFor(member: Member | null): OrganizationKind | null {
  if (!member || member.status !== 'active' || member.active === false) return null
  return member.organizationKind === 'enterprise' || member.organizationKind === 'department'
    ? member.organizationKind
    : null
}

export function createAuth() {
  const member = { value: null as Member | null }
  const state = { value: 'loading' as AuthState }

  const clear = () => {
    member.value = null
    state.value = 'unauthenticated'
  }

  const bootstrap = async () => {
    try {
      member.value = await api.me()
      state.value = workspaceFor(member.value) ? 'authenticated' : 'forbidden'
    } catch (error) {
      if ((error as { status?: number }).status === 403) state.value = 'forbidden'
      else clear()
    }
  }

  const login = async (loginId: string, password: string) => {
    const response = await api.login(loginId, password)
    member.value = response.member
    state.value = workspaceFor(member.value) ? 'authenticated' : 'forbidden'
    if (state.value !== 'authenticated') throw new Error('forbidden')
  }

  const logout = async () => {
    try { await api.logout() } finally { clear() }
  }

  return { member, state, bootstrap, login, logout, clear }
}
