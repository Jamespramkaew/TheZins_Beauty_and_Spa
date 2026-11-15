import axios from 'axios'
import { env } from '../config/env'

const axiosClient = axios.create({
  baseURL: env.apiBaseUrl,
  timeout: env.apiTimeout,
  // The backend authenticates with a session cookie (the_zins_session)
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Called when any request comes back 401 (session missing or expired).
// AuthProvider registers it to sign the user out and send them to the login page.
let unauthorizedHandler = null

export const setUnauthorizedHandler = (handler) => {
  unauthorizedHandler = handler
}

axiosClient.interceptors.response.use(
  (response) => response.data,
  (error) => {
    // Requests that handle 401 themselves (login, session check) opt out with `skipAuthRedirect`
    if (error.response?.status === 401 && !error.config?.skipAuthRedirect) {
      unauthorizedHandler?.()
    }
    return Promise.reject(error)
  },
)

export default axiosClient
