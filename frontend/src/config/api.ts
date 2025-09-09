interface ApiConfig {
  baseUrl: string
  timeout: number
}

// 默认配置
const defaultConfig: ApiConfig = {
  baseUrl: 'http://localhost:13544/api/v1',
  timeout: 10000,
}

const configs: Record<string, ApiConfig> = {
  development: {
    baseUrl: 'http://localhost:13544/api/v1',
    timeout: 10000,
  },
  production: {
    baseUrl: 'https://api.thepass.com/api/v1',
    timeout: 5000,
  },
  test: {
    baseUrl: 'http://localhost:13544/api/v1',
    timeout: 5000,
  },
}

// 获取当前环境，确保有fallback
const getEnvironment = (): string => {
  if (typeof import.meta !== 'undefined' && import.meta.env) {
    return import.meta.env.MODE || 'development'
  }
  return 'development'
}

const env = getEnvironment()

// 获取配置，确保总是有有效配置
const getApiConfig = (): ApiConfig => {
  let config = configs[env] || defaultConfig

  // 从环境变量覆盖baseUrl（如果存在）
  try {
    const envBaseURL = import.meta?.env?.VITE_API_BASE_URL
    if (envBaseURL && typeof envBaseURL === 'string') {
      config = { ...config, baseUrl: envBaseURL }
    }
  } catch (error) {
    console.warn('Failed to read environment variable VITE_API_BASE_URL:', error)
  }

  return config
}

export const apiConfig = getApiConfig()

// 调试信息（仅在开发环境）
if (env === 'development') {
  console.log('Environment MODE:', env)
  console.log('Available environments:', Object.keys(configs))
  console.log('Final API Config:', apiConfig)
}
