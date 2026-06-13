<template>
  <div class="page-grid two">
    <el-card shadow="never">
      <template #header>个人消息验证</template>
      <el-descriptions :column="1" border>
        <el-descriptions-item label="用户名">{{ profile.userName || auth.name }}</el-descriptions-item>
        <el-descriptions-item label="Token">
          <div class="copy-row">
            <el-input :model-value="profile.token" readonly />
            <el-button @click="copy(profile.token)">复制</el-button>
          </div>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
    <el-card shadow="never">
      <template #header>发送示例</template>
      <div class="send-examples">
        <section class="example-block">
          <div class="example-head">
            <span>GET API</span>
            <el-button size="small" @click="copy(sendUrl)">复制</el-button>
          </div>
          <el-input :model-value="sendUrl" readonly />
        </section>
        <section class="example-block">
          <div class="example-head">
            <span>POST API</span>
            <el-button size="small" @click="copy(postUrl)">复制</el-button>
          </div>
          <el-input :model-value="postUrl" readonly />
        </section>
        <section class="example-block">
          <div class="example-head">
            <span>POST JSON</span>
            <el-button size="small" @click="copy(postBody)">复制</el-button>
          </div>
          <el-input :model-value="postBody" type="textarea" :rows="8" readonly />
        </section>
        <span class="muted send-example-note">通过 token 发送给当前用户启用的所有消息通道。</span>
      </div>
    </el-card>
    <el-card shadow="never">
      <template #header>Bark 扫码绑定</template>
      <div class="bark-bind">
        <div class="qr-box">
          <img v-if="barkQr" :src="barkQr" alt="Bark 绑定二维码" />
        </div>
        <el-form label-width="88px">
          <el-form-item label="绑定地址">
            <div class="copy-row">
              <el-input :model-value="barkUrl" readonly />
              <el-button @click="copy(barkUrl)">复制</el-button>
            </div>
          </el-form-item>
          <el-form-item label="说明">
            <span class="muted">使用 Bark App 扫码或打开绑定地址，设备会注册为 Bark 通道。</span>
          </el-form-item>
        </el-form>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import QRCode from 'qrcode'

import { getSetting } from '@/api/setting'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const profile = ref({})
const sendUrl = computed(() => `${window.location.origin}/api/send?token=${profile.value.token || ''}&title=标题&body=内容`)
const postUrl = computed(() => `${window.location.origin}/api/send`)
const postBody = computed(() => JSON.stringify({
  token: profile.value.token || '',
  title: '标题',
  data: '第一行\\n第二行'
}, null, 2))
const barkUrl = computed(() => `${window.location.origin}/Register?act=${encodeURIComponent(profile.value.token || '')}`)
const barkQr = ref('')

onMounted(async () => {
  profile.value = await getSetting()
  barkQr.value = await QRCode.toDataURL(barkUrl.value, {
    width: 180,
    margin: 1
  })
})

async function copy(text) {
  await navigator.clipboard.writeText(text || '')
  ElMessage.success('已复制')
}
</script>
