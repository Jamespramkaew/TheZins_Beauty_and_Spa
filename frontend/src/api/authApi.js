import axiosClient from './axiosClient'

// User shape (login → me): { user_id, username, role: 'admin' | 'branch', branch_id: string | null }
export const authApi = {
  // Sets the session cookie. A 401 here means wrong credentials, not an expired session.
  login: ({ username, password }) =>
    axiosClient.post('/auth/login', { username, password }, { skipAuthRedirect: true }),
  // Who is logged in? 401 = no valid session.
  me: () => axiosClient.get('/auth/me', { skipAuthRedirect: true }),
  logout: () => axiosClient.post('/auth/logout'),
}
