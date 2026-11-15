import { useCallback, useEffect, useMemo, useState } from 'react'
import { authApi } from '../../api/authApi'
import { setUnauthorizedHandler } from '../../api/axiosClient'
import { AuthContext } from './AuthContext'

// The session cookie is HttpOnly, so the only way to know who is logged in is to ask
// the backend (GET /api/auth/me). The backend's AuthMiddleware remains the real guard.
function AuthProvider({ children }) {
  // signedOutReason: 'expired' (a request got 401) | 'signed-out' (pressed logout) | null
  const [auth, setAuth] = useState({ status: 'checking', user: null, signedOutReason: null })

  useEffect(() => {
    let active = true
    authApi
      .me()
      .then((response) => active && setAuth({ status: 'authenticated', user: response.data, signedOutReason: null }))
      .catch(() => active && setAuth({ status: 'guest', user: null, signedOutReason: null }))
    return () => {
      active = false
    }
  }, [])

  useEffect(() => {
    setUnauthorizedHandler(() =>
      setAuth((current) =>
        current.status === 'authenticated' ? { status: 'guest', user: null, signedOutReason: 'expired' } : current,
      ),
    )
    return () => setUnauthorizedHandler(null)
  }, [])

  const signIn = useCallback(async (credentials) => {
    await authApi.login(credentials)
    const response = await authApi.me()
    setAuth({ status: 'authenticated', user: response.data, signedOutReason: null })
    return response.data
  }, [])

  const signOut = useCallback(async () => {
    try {
      await authApi.logout()
    } catch {
      // Sign out locally even if the server is unreachable
    }
    setAuth({ status: 'guest', user: null, signedOutReason: 'signed-out' })
  }, [])

  const value = useMemo(() => ({ ...auth, signIn, signOut }), [auth, signIn, signOut])

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export default AuthProvider
