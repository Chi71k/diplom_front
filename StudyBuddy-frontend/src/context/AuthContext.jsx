import { createContext, useState, useCallback } from 'react'
import { getToken } from '../api'

const AuthContext = createContext(null)

function decodeJwtRole(token) {
  if (!token) return null
  try {
    const payload = JSON.parse(atob(token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')))
    return payload.role || null
  } catch {
    return null
  }
}

export function AuthProvider({ children }) {
  const [token, setTokenState] = useState(() => getToken())
  const [profile, setProfile] = useState(null)

  const setToken = useCallback((newToken) => {
    if (newToken) {
      localStorage.setItem('accessToken', newToken)
      setTokenState(newToken)
    } else {
      localStorage.removeItem('accessToken')
      localStorage.removeItem('refreshToken')
      setTokenState(null)
      setProfile(null)
    }
  }, [])

  const role = decodeJwtRole(token)

  const value = {
    token,
    setToken,
    profile,
    setProfile,
    isAuthenticated: !!token,
    role,
    isAdmin: role === 'admin',
  }
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export { AuthContext }
