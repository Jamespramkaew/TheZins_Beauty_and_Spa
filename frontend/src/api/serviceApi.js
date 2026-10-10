import axiosClient from './axiosClient'

export const serviceApi = {
  getAll: () => axiosClient.get('/services'),
  getById: (id) => axiosClient.get(`/services/${id}`),
}
