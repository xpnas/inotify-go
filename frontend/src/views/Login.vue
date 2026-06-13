<template>
  <main class="login-page">
    <section class="login-panel">
      <div class="login-copy">
        <p class="login-tagline">Notification Gateway</p>
        <h1>Inotify</h1>
        <p>统一配置通知通道、发送密钥和系统参数，让消息触达每一个终端。</p>
        <ul class="login-features">
          <li>支持邮件、Telegram、企业微信、钉钉、飞书等多渠道</li>
          <li>统一 API，一次发送覆盖所有通道</li>
          <li>Bark 扫码一键绑定 iOS 设备</li>
          <li>完整消息发送历史记录</li>
        </ul>
      </div>
      <el-form ref="formRef" :model="form" :rules="rules" class="login-form" @keyup.enter="submit">
        <p class="login-form-title">欢迎回来</p>
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="用户名" size="large" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" placeholder="密码" type="password" show-password size="large" />
        </el-form-item>
        <el-button type="primary" size="large" :loading="loading" @click="submit">登录</el-button>
        <el-button v-if="githubEnabled" class="github-login" size="large" :loading="githubLoading" @click="loginWithGithub">
          GitHub 登录
        </el-button>
        <el-button v-if="weixinQrEnabled" class="github-login" size="large" :loading="weixinQrLoading" @click="loginWithWeixinQr">
          微信扫码登录
        </el-button>
      </el-form>
    </section>
  </main>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { githubLogin, weixinQrLogin } from '@/api/user'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const formRef = ref()
const loading = ref(false)
const githubLoading = ref(false)
const githubEnabled = ref(false)
const weixinQrLoading = ref(false)
const weixinQrEnabled = ref(false)
const form = reactive({ username: 'admin', password: '' })
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

onMounted(async () => {
  const [ghEnabled, wxEnabled] = await Promise.all([
    fetch('/api/oauth/GithubEnable').then(r => r.json()).then(j => j?.data === true).catch(() => false),
    fetch('/api/oauth/WeixinQrEnable').then(r => r.json()).then(j => j?.data === true).catch(() => false)
  ])
  githubEnabled.value = Boolean(ghEnabled)
  weixinQrEnabled.value = Boolean(wxEnabled)

  const code = route.query.code
  const state = route.query.state
  if (code && state === 'inotify_weixin_login') {
    weixinQrLoading.value = true
    try {
      const data = await weixinQrLogin({ code })
      auth.acceptLogin(data)
      router.replace('/')
    } catch {
      weixinQrLoading.value = false
    }
    return
  }
  if (code) {
    githubLoading.value = true
    try {
      const data = await githubLogin(code)
      auth.acceptLogin(data)
      router.push('/')
    } finally {
      githubLoading.value = false
    }
  }
})

async function submit() {
  await formRef.value.validate()
  loading.value = true
  try {
    await auth.login(form)
    await auth.loadProfile()
    router.push(route.query.redirect || '/')
  } finally {
    loading.value = false
  }
}

async function loginWithGithub() {
  githubLoading.value = true
  try {
    const url = await githubLogin('')
    window.location.href = url
  } finally {
    githubLoading.value = false
  }
}

async function loginWithWeixinQr() {
  weixinQrLoading.value = true
  try {
    const redirectUri = `${window.location.origin}/login`
    const url = await weixinQrLogin({ redirectUri })
    window.location.href = url
  } finally {
    weixinQrLoading.value = false
  }
}
</script>
