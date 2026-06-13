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

      <!-- 二维码视图 -->
      <div v-if="showQr" class="login-form login-qr-view">
        <p class="login-form-title">微信扫码登录</p>
        <div v-if="weixinQrLoading" class="weixin-qr-loading">
          <el-icon class="is-loading" :size="32"><Loading /></el-icon>
          <span>正在加载二维码…</span>
        </div>
        <div id="weixin-qr-container" ref="qrContainer" />
        <p class="weixin-qr-tip muted">使用微信或企业微信扫码登录</p>
        <el-button size="large" style="width:100%;margin-top:4px;" @click="backToPassword">返回密码登录</el-button>
      </div>

      <!-- 密码登录视图 -->
      <el-form v-else ref="formRef" :model="form" :rules="rules" class="login-form" @keyup.enter="submit">
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
        <el-button v-if="weixinQrEnabled" class="github-login" size="large" @click="openWeixinQr">
          微信扫码登录
        </el-button>
      </el-form>
    </section>
  </main>
</template>

<script setup>
import { nextTick, onMounted, reactive, ref } from 'vue'
import { Loading } from '@element-plus/icons-vue'
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
const showQr = ref(false)
const qrContainer = ref(null)
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
    showQr.value = true
    weixinQrLoading.value = true
    try {
      const data = await weixinQrLogin({ code })
      auth.acceptLogin(data)
      router.replace('/')
    } catch {
      weixinQrLoading.value = false
      showQr.value = false
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

async function openWeixinQr() {
  showQr.value = true
  weixinQrLoading.value = true
  await nextTick()
  try {
    const redirectUri = `${window.location.origin}/login`
    const qrUrl = await weixinQrLogin({ redirectUri })
    const parsed = new URL(qrUrl)
    const appid = parsed.searchParams.get('appid') || ''
    const agentid = parsed.searchParams.get('agentid') || ''
    const state = parsed.searchParams.get('state') || ''
    await loadWwLoginSdk()
    if (qrContainer.value) qrContainer.value.innerHTML = ''
    // eslint-disable-next-line no-undef
    new WwLogin({
      id: 'weixin-qr-container',
      appid,
      agentid,
      redirect_uri: encodeURIComponent(redirectUri),
      state,
      href: '',
      lang: 'zh',
    })
  } catch {
    showQr.value = false
  } finally {
    weixinQrLoading.value = false
  }
}

function backToPassword() {
  showQr.value = false
  if (qrContainer.value) qrContainer.value.innerHTML = ''
}

function loadWwLoginSdk() {
  return new Promise((resolve, reject) => {
    if (window.WwLogin) { resolve(); return }
    const s = document.createElement('script')
    s.src = 'https://wwcdn.weixin.qq.com/node/wework/wwopen/js/wwLogin-1.2.7.js'
    s.onload = resolve
    s.onerror = reject
    document.head.appendChild(s)
  })
}
</script>
