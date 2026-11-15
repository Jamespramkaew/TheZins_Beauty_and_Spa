// Backend error body: { success: false, message, data: [{ error: { type, details } }] }
export function readApiError(error) {
  const body = error?.response?.data
  return {
    status: error?.response?.status,
    message: body?.message ?? '',
    detail: body?.data?.[0]?.error?.details ?? '',
  }
}
