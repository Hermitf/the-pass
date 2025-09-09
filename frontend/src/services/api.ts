import axios from 'axios'
import { apiConfig } from '../config/api'
import { HTTP_STATUS } from '../config/constants'
import type { LoginRequest, LoginResponse } from '../config/types'

// 创建axios实例
const api = axios.create({
  baseURL: apiConfig.baseUrl,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器 - 添加认证token
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器 - 处理错误
api.interceptors.response.use(
  response => response,
  error => {
    if (error.response?.status === HTTP_STATUS.UNAUTHORIZED) {
      // TODO Token无效，清除本地存储
      // localStorage.removeItem('token');
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default api

export const authAPI = {
  // 用户登录
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    const response = await api.post<LoginResponse>('/users/login', data)
    return response.data
  },
}
