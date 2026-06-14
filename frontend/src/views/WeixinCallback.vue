<template>
  <main class="oauth-callback">
    <section class="oauth-callback-panel">
      <el-icon class="oauth-callback-icon" :class="status">
        <Loading v-if="status === 'loading'" />
        <CircleCheck v-else-if="status === 'success'" />
        <CircleClose v-else />
      </el-icon>
      <h1>{{ title }}</h1>
      <p>{{ message }}</p>
      <el-button v-if="status === 'error'" type="primary" @click="goBack">返回</el-button>
    </section>
  </main>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { CircleCheck, CircleClose, Loading } from '@element-plus/icons-vue'
import { useRoute, useRouter } from 'vue-router'

import { weixinQrLogin } from '@/api/user'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const status = ref('loading')
const mode = computed(() => route.query.state === 'inotify_weixin_bind' ? 'bind' : 'login')
const title = computed(() => {
  if (status.value === 'loading') return mode.value === 'bind' ? '正在绑定企业微信' : '正在完成企业微信登录'
  if (status.value === 'success') return mode.value === 'bind' ? '绑定成功' : '登录成功'
  return mode.value === 'bind' ? '绑定失败' : '登录失败'
})
const message = ref('请稍候，正在和企业微信确认授权结果。')

onMounted(async () => {
  const code = route.query.code
  if (!code) {
    status.value = 'error'
    message.value = '缺少企业微信授权 code'
    return
  }

  try {
    const redirectUri = `${window.location.origin}/oauth/weixin/callback`
    if (mode.value === 'bind') {
      if (!auth.token) throw new Error('请先登录后再绑定企业微信')
      const data = await auth.weixinQrBind({ code, redirectUri })
      status.value = 'success'
      message.value = `已绑定企业微信账号：${data.weixinId}`
      setTimeout(() => router.replace('/settingpro/thirdparty'), 900)
      return
    }
    const data = await weixinQrLogin({ code })
    auth.acceptLogin(data)
    status.value = 'success'
    message.value = '授权完成，正在进入系统。'
    setTimeout(() => router.replace('/'), 600)
  } catch (error) {
    status.value = 'error'
    message.value = error?.message || '企业微信授权处理失败'
  }
})

function goBack() {
  router.replace(auth.token ? '/settingpro/thirdparty' : '/login')
}
</script>
