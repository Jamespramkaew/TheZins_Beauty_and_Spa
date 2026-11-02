import { useContext } from 'react'
import { AuthContext } from '../features/auth/AuthContext'

// Usage: const { status, user, signIn, signOut } = useAuth()
//   status: 'checking' | 'authenticated' | 'guest'
export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) throw new Error('useAuth must be used within AuthProvider')
  return context
}
