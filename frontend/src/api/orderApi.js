import axiosClient from './axiosClient'

// Branch accounts only ever get their own branch's orders; admin sees every branch.
export const orderApi = {
  // params: { customer_id, from, to, status: paid|partial|unpaid, page, page_size }
  getAll: (params) => axiosClient.get('/orders', { params }),
  // Order + customer + branch + items (with course/stock) + payments + live total/paid/outstanding/status
  getById: (id) => axiosClient.get(`/orders/${id}`),
  // body: { method: cash|transfer|card, amount: "500.00", reference } → the updated order
  addPayment: (id, body) => axiosClient.post(`/orders/${id}/payments`, body),
}
