<template>
  <el-card shadow="never">
    <template #header>Bark 绑定</template>
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
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import QRCode from 'qrcode'

import { getSetting } from '@/api/setting'

const profile = ref({})
const barkQr = ref('')
const barkUrl = computed(() => `${window.location.origin}/Register?act=${encodeURIComponent(profile.value.token || '')}`)

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
