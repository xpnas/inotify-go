import axios from 'axios'
import { ElMessage } from 'element-plus'

import { useAuthStore } from '@/stores/auth'

const service = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || '/api',
  timeout: 15000
})

service.interceptors.request.use((config) => {
  const auth = useAuthStore()
  if (auth.token) {
    config.headers['X-Token'] = auth.token
  }
  return config
})

service.interceptors.response.use(
  (response) => {
    const body = response.data
    if (!body || typeof body.code === 'undefined') {
      return body
    }
    if (body.code !== 20000) {
      ElMessage.error(body.msg || '请求失败')
      return Promise.reject(new Error(body.msg || '请求失败'))
    }
    return body.data
  },
  (error) => {
    if (!error.config?.silentError) {
      ElMessage.error(error.message || '网络异常')
    }
    return Promise.reject(error)
  }
)

export default service
