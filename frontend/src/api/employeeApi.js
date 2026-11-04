import axiosClient from './axiosClient'

export const employeeApi = {
  // Active employees. Admin may pass branch_id; branch accounts always get their own branch.
  getAll: (params) => axiosClient.get('/employees', { params }),
}
