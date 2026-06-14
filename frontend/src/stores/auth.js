import { defineStore } from 'pinia'

import { getInfo, githubBind, githubLogin, githubUnbind, login, logout, weixinQrBind, weixinQrUnbind } from '@/api/user'

const TOKEN_KEY = 'inotify-token'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem(TOKEN_KEY) || '',
    name: '',
    role: '',
    avatar: '',
    email: '',
    githubLogin: '',
    githubId: 0,
    weixinId: ''
  }),
  getters: {
    isSystem: (state) => state.role === 'System' || state.role === 'system'
  },
  actions: {
    async login(payload) {
      const data = await login(payload)
      this.token = data.token
      this.name = data.name
      this.role = data.role
      localStorage.setItem(TOKEN_KEY, data.token)
      return data
    },
    async loadProfile() {
      if (!this.token) return null
      const data = await getInfo()
      this.name = data.name
      this.role = data.role || (Array.isArray(data.roles) ? data.roles[0] : '')
      this.avatar = data.avatar || ''
      this.email = data.email || ''
      this.githubLogin = data.githubLogin || ''
      this.githubId = data.githubId || 0
      this.weixinId = data.weixinId || ''
      return data
    },
    acceptLogin(data) {
      this.token = data.token
      this.name = data.name
      this.role = data.role
      this.avatar = data.avatar || ''
      this.email = data.email || ''
      this.githubLogin = data.githubLogin || ''
      this.githubId = data.githubId || 0
      this.weixinId = data.weixinId || ''
      localStorage.setItem(TOKEN_KEY, data.token)
    },
    async githubLogin(params) {
      const data = await githubLogin(params)
      this.acceptLogin(data)
      return data
    },
    async githubBind(params) {
      const data = await githubBind(params)
      this.githubLogin = data.githubLogin || ''
      this.githubId = data.githubId || 0
      return data
    },
    async githubUnbind() {
      await githubUnbind()
      this.githubLogin = ''
      this.githubId = 0
    },
    async weixinQrBind(params) {
      const data = await weixinQrBind(params)
      this.weixinId = data.weixinId || ''
      return data
    },
    async weixinQrUnbind() {
      await weixinQrUnbind()
      this.weixinId = ''
    },
    async logout() {
      try {
        if (this.token) await logout()
      } finally {
        this.clear()
      }
    },
    clear() {
      this.token = ''
      this.name = ''
      this.role = ''
      this.avatar = ''
      this.email = ''
      this.githubLogin = ''
      this.githubId = 0
      this.weixinId = ''
      localStorage.removeItem(TOKEN_KEY)
    }
  }
})
