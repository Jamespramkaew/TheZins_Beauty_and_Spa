import axiosClient from './axiosClient'

// Branch accounts are always locked to their own branch by the backend; branch_id only matters for admin.
export const stockLogApi = {
  // params: { branch_id, search, category, low_stock, active, page, page_size }
  getQuantities: (params) => axiosClient.get('/stock-logs/quantities', { params }),
  // params: { stock_id, branch_id, log_type, direction, from, to, page, page_size }
  getLogs: (params) => axiosClient.get('/stock-logs', { params }),
  // body: { type, branch_id, stock_id, employee_id, qty, note }
  addMovement: (body) => axiosClient.post('/stock-logs', body),
  // body: { stock_id, from_branch_id, to_branch_id, employee_id, qty }
  transfer: (body) => axiosClient.post('/stock-logs/transfer', body),
}
