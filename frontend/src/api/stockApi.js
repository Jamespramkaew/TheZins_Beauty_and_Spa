import axiosClient from './axiosClient'

export const stockApi = {
  getAll: () => axiosClient.get('/stocks'),
}
