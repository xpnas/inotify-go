<template>
  <div class="shell">
    <aside class="sidebar">
      <div class="brand">Inotify</div>
      <el-menu :default-active="$route.path" router>
        <el-menu-item
          v-for="item in visibleMenu"
          :key="item.path"
          :index="item.path"
        >
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.title }}</span>
        </el-menu-item>
      </el-menu>
      <a class="source-link" href="https://github.com/xpnas/Inotify" target="_blank" rel="noreferrer">
        源码
      </a>
    </aside>
    <main class="main">
      <header class="topbar">
        <div>
          <h1>{{ currentTitle }}</h1>
          <p>消息通知管理后台</p>
        </div>
        <el-dropdown trigger="click" @command="handleCommand">
          <button class="user-button">
            <el-avatar :size="32" :src="auth.avatar">{{ avatarText }}</el-avatar>
            <span>{{ auth.name || '用户' }}</span>
            <el-tag size="small" effect="plain">{{ auth.role || 'Role' }}</el-tag>
          </button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </header>
      <section class="content">
        <router-view />
      </section>
    </main>
  </div>
</template>

<script setup>
import {
  Bell,
  Clock,
  Key,
  Link,
  Lock,
  Monitor,
  Setting,
  User
} from '@element-plus/icons-vue'
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { menuRoutes } from '@/router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const icons = [Key, Bell, Clock, Lock, Monitor, User, Setting, Link]

const visibleMenu = computed(() => {
  const children = menuRoutes[0].children
  return children
    .filter((item) => !item.meta.roles || item.meta.roles.includes(auth.role))
    .map((item, index) => ({
      path: item.path,
      title: item.meta.title,
      icon: icons[index] || Setting
    }))
})

const currentTitle = computed(() => route.meta.title || 'Inotify')
const avatarText = computed(() => (auth.name || 'I').slice(0, 1).toUpperCase())

async function handleCommand(command) {
  if (command === 'logout') {
    await auth.logout()
    router.push('/login')
  }
}
</script>
