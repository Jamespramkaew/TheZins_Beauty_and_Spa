import axiosClient from './axiosClient'

export const courseApi = {
  getAll: (params) => axiosClient.get('/courses', { params }),
  create: (body) => axiosClient.post('/courses', body),
  update: (id, body) => axiosClient.patch(`/courses/${id}`, body),
  remove: (id) => axiosClient.delete(`/courses/${id}`),
}
