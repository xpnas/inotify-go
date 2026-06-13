import { defineStore } from 'pinia'

import { getInfo, githubLogin, login, logout } from '@/api/user'

const TOKEN_KEY = 'inotify-token'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem(TOKEN_KEY) || '',
    name: '',
    role: '',
    avatar: '',
    email: ''
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
      return data
    },
    acceptLogin(data) {
      this.token = data.token
      this.name = data.name
      this.role = data.role
      this.avatar = data.avatar || ''
      this.email = data.email || ''
      localStorage.setItem(TOKEN_KEY, data.token)
    },
    async githubLogin(code) {
      const data = await githubLogin(code)
      this.acceptLogin(data)
      return data
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
      localStorage.removeItem(TOKEN_KEY)
    }
  }
})
