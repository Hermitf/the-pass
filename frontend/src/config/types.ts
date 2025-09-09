// 登录请求和响应类型
export interface LoginRequest {
  login_info: string
  password: string
  login_type: string
}

export interface LoginResponse {
  token: string
  user: {
    id: string
    email: string
    username: string
    phone: string
  }
}
