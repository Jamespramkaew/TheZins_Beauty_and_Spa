import axiosClient from './axiosClient'

// Branch accounts only ever get their own branch's orders; admin sees every branch.
export const orderApi = {
  // params: { customer_id, from, to, status: paid|partial|unpaid, page, page_size }
  getAll: (params) => axiosClient.get('/orders', { params }),
}
