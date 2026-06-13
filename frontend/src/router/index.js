import { createRouter, createWebHistory } from 'vue-router'

import ShellLayout from '@/layout/ShellLayout.vue'
import { useAuthStore } from '@/stores/auth'

export const menuRoutes = [
  {
    path: '/',
    redirect: '/settingpro/sendkey',
    component: ShellLayout,
    children: [
      {
        path: '/settingpro/sendkey',
        name: 'SendKey',
        component: () => import('@/views/SendKey.vue'),
        meta: { title: '消息验证', roles: ['User', 'System'] }
      },
      {
        path: '/settingpro/sendmethods',
        name: 'SendAuths',
        component: () => import('@/views/SendAuths.vue'),
        meta: { title: '消息通道', roles: ['User', 'System'] }
      },
      {
        path: '/settingpro/history',
        name: 'MessageHistory',
        component: () => import('@/views/MessageHistory.vue'),
        meta: { title: '历史记录', roles: ['User', 'System'] }
      },
      {
        path: '/settingpro/oauthsetting',
        name: 'Password',
        component: () => import('@/views/Password.vue'),
        meta: { title: '重置密码', roles: ['User', 'System'] }
      },
      {
        path: '/settingsys/systeminfo',
        name: 'SystemInfo',
        component: () => import('@/views/settings/SystemInfo.vue'),
        meta: { title: '系统状态', roles: ['System'] }
      },
      {
        path: '/settingsys/usermanager',
        name: 'Users',
        component: () => import('@/views/settings/Users.vue'),
        meta: { title: '用户管理', roles: ['System'] }
      },
      {
        path: '/settingsys/jwt',
        name: 'JWT',
        component: () => import('@/views/settings/Jwt.vue'),
        meta: { title: 'JWT 参数', roles: ['System'] }
      },
      {
        path: '/settingsys/systemglobal',
        name: 'Global',
        component: () => import('@/views/settings/Global.vue'),
        meta: { title: '全局参数', roles: ['System'] }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'Login', component: () => import('@/views/Login.vue'), meta: { public: true } },
    ...menuRoutes,
    { path: '/:pathMatch(.*)*', name: 'NotFound', component: () => import('@/views/NotFound.vue'), meta: { public: true } }
  ]
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (to.meta.public) {
    if (to.path === '/login' && auth.token) return '/'
    return true
  }
  if (!auth.token) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if (!auth.role) {
    try {
      await auth.loadProfile()
    } catch {
      auth.clear()
      return { path: '/login', query: { redirect: to.fullPath } }
    }
  }
  const roles = to.meta.roles || []
  if (roles.length && !roles.includes(auth.role)) {
    return '/'
  }
  return true
})

export default router
