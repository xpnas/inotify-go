<template>
  <div class="stack">
    <el-card shadow="never" class="form-card">
      <template #header>GitHub 登录绑定</template>
      <el-form label-width="120px">
        <el-form-item label="绑定状态">
          <span>{{ auth.githubLogin ? `已绑定：${auth.githubLogin}` : '未绑定' }}</span>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="githubBinding" @click="bindGithub">
            {{ auth.githubLogin ? '重新绑定 GitHub' : '绑定 GitHub' }}
          </el-button>
          <el-popconfirm v-if="auth.githubLogin" title="确认解除 GitHub 绑定？" @confirm="unbindGithub">
            <template #reference>
              <el-button :loading="githubUnbinding">解除绑定</el-button>
            </template>
          </el-popconfirm>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="form-card">
      <template #header>企业微信登录绑定</template>
      <el-form label-width="120px">
        <el-form-item label="绑定状态">
          <span>{{ auth.weixinId ? `已绑定：${auth.weixinId}` : '未绑定' }}</span>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="weixinBinding" @click="bindWeixin">
            {{ auth.weixinId ? '重新绑定企业微信' : '绑定企业微信' }}
          </el-button>
          <el-popconfirm v-if="auth.weixinId" title="确认解除企业微信绑定？" @confirm="unbindWeixin">
            <template #reference>
              <el-button :loading="weixinUnbinding">解除绑定</el-button>
            </template>
          </el-popconfirm>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'

import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const githubBinding = ref(false)
const weixinBinding = ref(false)
const githubUnbinding = ref(false)
const weixinUnbinding = ref(false)

onMounted(() => {
  auth.loadProfile()
})

async function bindGithub() {
  githubBinding.value = true
  try {
    const redirectUri = `${window.location.origin}/oauth/github/callback`
    const url = await auth.githubBind({ redirectUri })
    window.location.href = url
  } finally {
    githubBinding.value = false
  }
}

async function bindWeixin() {
  weixinBinding.value = true
  try {
    const redirectUri = `${window.location.origin}/oauth/weixin/callback`
    const url = await auth.weixinQrBind({ redirectUri })
    window.location.href = url
  } finally {
    weixinBinding.value = false
  }
}

async function unbindGithub() {
  githubUnbinding.value = true
  try {
    await auth.githubUnbind()
  } finally {
    githubUnbinding.value = false
  }
}

async function unbindWeixin() {
  weixinUnbinding.value = true
  try {
    await auth.weixinQrUnbind()
  } finally {
    weixinUnbinding.value = false
  }
}
</script>
