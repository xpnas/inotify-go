<template>
  <main class="login-page">
    <section class="login-panel">
      <div class="login-copy">
        <h1>Inotify</h1>
        <p>统一配置通知通道、发送密钥和系统参数。</p>
      </div>
      <el-form ref="formRef" :model="form" :rules="rules" class="login-form" @keyup.enter="submit">
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
      </el-form>
    </section>
  </main>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getGithubEnable, githubLogin } from '@/api/user'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const formRef = ref()
const loading = ref(false)
const githubLoading = ref(false)
const githubEnabled = ref(false)
const form = reactive({ username: 'admin', password: '' })
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

onMounted(async () => {
  githubEnabled.value = Boolean(await getGithubEnable().catch(() => false))
  if (route.query.code) {
    githubLoading.value = true
    try {
      const data = await githubLogin(route.query.code)
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
</script>
