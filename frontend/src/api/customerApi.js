import axiosClient from './axiosClient'

export const customerApi = {
  // params: { q, tier: general|vip, page, page_size } — empty tier = every tier
  search: (params) => axiosClient.get('/customers', { params }),
  getById: (id) => axiosClient.get(`/customers/${id}`),
  create: (body) => axiosClient.post('/customers', body),
  update: (id, body) => axiosClient.patch(`/customers/${id}`, body),
  changeTier: (id, tier) => axiosClient.post(`/customers/${id}/tier`, { tier }),
  getTierChanges: (id) => axiosClient.get(`/customers/${id}/tier-changes`),
}
