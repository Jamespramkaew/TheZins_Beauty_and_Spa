import axiosClient from './axiosClient'

export const branchApi = {
  // Every active branch (id, code, name) — branch accounts need the others as transfer destinations
  getAll: () => axiosClient.get('/branches'),
}
