<template>
  <div class="page-grid two">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>个人消息验证</span>
          <el-button size="small" type="primary" :loading="testingChannels" @click="testAllChannels">测试所有通道</el-button>
        </div>
      </template>
      <el-descriptions :column="1" border>
        <el-descriptions-item label="用户名">{{ profile.userName || auth.name }}</el-descriptions-item>
        <el-descriptions-item label="Token">
          <div class="copy-row">
            <el-input :model-value="profile.token" readonly />
            <el-button @click="copy(profile.token)">复制</el-button>
          </div>
        </el-descriptions-item>
      </el-descriptions>
      <div v-if="channelResults.length" class="channel-results">
        <div
          v-for="r in channelResults"
          :key="r.name"
          class="channel-result-item"
          :class="r.status"
        >
          <span class="channel-result-icon">
            <span v-if="r.status === 'ok'">✓</span>
            <span v-else-if="r.status === 'fail'">✗</span>
            <span v-else class="loading-dot">…</span>
          </span>
          <span class="channel-result-name">{{ r.name }}</span>
          <span class="channel-result-type muted">{{ r.typeName }}</span>
          <span v-if="r.msg" class="channel-result-msg">{{ r.msg }}</span>
        </div>
      </div>
    </el-card>
    <el-card shadow="never">
      <template #header>发送示例</template>
      <div class="send-examples">
        <section class="example-block">
          <div class="example-head">
            <span>GET API</span>
            <div class="example-actions">
              <el-button size="small" @click="copy(sendUrl)">复制</el-button>
              <el-button size="small" @click="resetExamples">还原</el-button>
              <el-button size="small" type="primary" :loading="testingGet" @click="testGet">测试</el-button>
            </div>
          </div>
          <el-input v-model="sendUrl" />
          <div v-if="getResult" class="test-result" :class="getResult.ok ? 'ok' : 'fail'">
            {{ getResult.ok ? '✓ 发送成功' : '✗ 发送失败' }}{{ getResult.msg ? '：' + getResult.msg : '' }}
          </div>
        </section>
        <section class="example-block">
          <div class="example-head">
            <span>POST API</span>
            <div class="example-actions">
              <el-button size="small" @click="copy(postUrl)">复制</el-button>
              <el-button size="small" type="primary" :loading="testingPost" @click="testPost">测试</el-button>
            </div>
          </div>
          <el-input v-model="postUrl" />
          <div v-if="postResult" class="test-result" :class="postResult.ok ? 'ok' : 'fail'">
            {{ postResult.ok ? '✓ 发送成功' : '✗ 发送失败' }}{{ postResult.msg ? '：' + postResult.msg : '' }}
          </div>
        </section>
        <section class="example-block">
          <div class="example-head">
            <span>POST JSON</span>
            <el-button size="small" @click="copy(postBody)">复制</el-button>
          </div>
          <el-input v-model="postBody" type="textarea" :rows="8" />
        </section>
        <span class="muted send-example-note">通过 token 发送给当前用户启用的所有消息通道。</span>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'

import { getSendAuths, getSendTemplates, getSetting } from '@/api/setting'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const profile = ref({})
const testingGet = ref(false)
const testingPost = ref(false)
const getResult = ref(null)
const postResult = ref(null)
const testingChannels = ref(false)
const channelResults = ref([])
const templates = ref([])
const sendUrl = ref('')
const postUrl = ref('')
const postBody = ref('')

onMounted(async () => {
  const [setting, tpls] = await Promise.all([getSetting(), getSendTemplates()])
  profile.value = setting
  templates.value = tpls || []
  resetExamples()
})

function resetExamples() {
  sendUrl.value = `${window.location.origin}/api/send?token=${profile.value.token || ''}&title=标题&body=内容`
  postUrl.value = `${window.location.origin}/api/send`
  postBody.value = JSON.stringify({
    token: profile.value.token || '',
    title: '标题',
    data: '第一行\\n第二行'
  }, null, 2)
}

async function testAllChannels() {
  testingChannels.value = true
  channelResults.value = []
  try {
    const auths = await getSendAuths()
    const active = (auths || []).filter(a => a.active)
    if (!active.length) {
      ElMessage.warning('没有已启用的消息通道')
      return
    }
    channelResults.value = active.map(a => ({
      name: a.name,
      key: a.key,
      typeName: templates.value.find(t => t.key === a.templateID)?.name || a.templateID,
      status: 'pending',
      msg: ''
    }))
    await Promise.all(
      channelResults.value.map(async (r) => {
        try {
          const url = `${window.location.origin}/api/send?key=${encodeURIComponent(r.key)}&title=测试消息&body=来自 Inotify 的通道测试`
          const resp = await fetch(url)
          const json = await resp.json().catch(() => ({}))
          r.status = (resp.ok && json.code === 20000) ? 'ok' : 'fail'
          r.msg = json.msg || ''
        } catch (e) {
          r.status = 'fail'
          r.msg = e.message
        }
      })
    )
  } finally {
    testingChannels.value = false
  }
}

async function copy(text) {
  await navigator.clipboard.writeText(text || '')
  ElMessage.success('已复制')
}

async function testGet() {
  testingGet.value = true
  getResult.value = null
  try {
    const resp = await fetch(sendUrl.value)
    const json = await resp.json().catch(() => ({}))
    getResult.value = { ok: resp.ok && json.code === 20000, msg: json.msg || '' }
  } catch (e) {
    getResult.value = { ok: false, msg: e.message }
  } finally {
    testingGet.value = false
  }
}

async function testPost() {
  testingPost.value = true
  postResult.value = null
  try {
    const resp = await fetch(postUrl.value, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: postBody.value
    })
    const json = await resp.json().catch(() => ({}))
    postResult.value = { ok: resp.ok && json.code === 20000, msg: json.msg || '' }
  } catch (e) {
    postResult.value = { ok: false, msg: e.message }
  } finally {
    testingPost.value = false
  }
}
</script>
