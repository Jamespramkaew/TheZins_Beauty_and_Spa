export const env = {
  appName: import.meta.env.VITE_APP_NAME,
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL,
  apiTimeout: Number(import.meta.env.VITE_API_TIMEOUT) || 10000,
}
