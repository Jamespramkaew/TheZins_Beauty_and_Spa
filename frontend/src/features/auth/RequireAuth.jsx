import { Navigate, useLocation } from 'react-router-dom'
import FullPageLoader from '../../components/FullPageLoader'
import { useAuth } from '../../hooks/useAuth'

// Wrap any route that needs a logged-in user. Guests go to the login page ("/"),
// which sends them back to the page they asked for after logging in.
function RequireAuth({ children }) {
  const { status, signedOutReason } = useAuth()
  const location = useLocation()

  if (status === 'checking') return <FullPageLoader />

  if (status === 'guest') {
    // After pressing logout, don't send the next person back to the previous user's page
    const state = signedOutReason === 'signed-out' ? undefined : { from: location, expired: signedOutReason === 'expired' }
    return <Navigate to="/" replace state={state} />
  }

  return children
}

export default RequireAuth
