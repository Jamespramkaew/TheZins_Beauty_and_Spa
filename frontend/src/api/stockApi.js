import axiosClient from './axiosClient'

export const stockApi = {
  getAll: () => axiosClient.get('/stocks'),
  // params: { status: open|closed|all, branch_id (admin only), page, page_size }
  getContainers: (id, params) => axiosClient.get(`/stocks/${id}/containers`, { params }),
  create: (body) => axiosClient.post('/stocks', body),
  update: (id, body) => axiosClient.patch(`/stocks/${id}`, body),
  remove: (id) => axiosClient.delete(`/stocks/${id}`),
}
