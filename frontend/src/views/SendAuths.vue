<template>
  <div class="stack">
    <div class="toolbar">
      <el-button type="primary" @click="openCreate">新增通道</el-button>
      <el-button @click="load">刷新</el-button>
    </div>
    <el-table :data="auths" border>
      <el-table-column prop="name" label="名称" min-width="140" />
      <el-table-column prop="templateID" label="模板" min-width="180">
        <template #default="{ row }">{{ templateName(row.templateID) }}</template>
      </el-table-column>
      <el-table-column prop="key" label="发送 Key" min-width="260" />
      <el-table-column prop="active" label="激活" width="96" align="center">
        <template #default="{ row }">
          <el-switch v-model="row.active" @change="(value) => toggle(row, value)" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="240" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link @click="resetKey(row)">重置 Key</el-button>
          <el-popconfirm title="确认删除这个通道？" @confirm="remove(row)">
            <template #reference>
              <el-button link type="danger">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog
      v-model="dialogVisible"
      :title="editing ? '编辑通道' : '新增通道'"
      width="min(92vw, 640px)"
      class="send-auth-dialog"
    >
      <el-form class="send-auth-form" label-width="112px">
        <el-form-item label="模板">
          <el-select v-model="form.templateID" filterable @change="handleTemplateChange">
            <el-option v-for="tpl in templates" :key="tpl.key" :label="tpl.name" :value="tpl.key">
              <span class="option-label">{{ tpl.name }}</span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="代理策略">
          <el-segmented v-model="form.config.ProxyMode" :options="proxyOptions" />
        </el-form-item>
        <el-form-item v-if="form.config.ProxyMode === 'custom'" label="代理地址">
          <el-input v-model="form.config.ProxyAddress" placeholder="http://127.0.0.1:7890" />
        </el-form-item>
        <template v-if="isWeixinScanTemplate">
          <el-form-item label="绑定方式">
            <div class="stack" style="width:100%;gap:10px;">
              <el-alert
                type="info"
                :closable="false"
                title="扫码后自动绑定到当前通道"
                description="若提示未配置，请先到 系统管理 > 企业微信扫码绑定配置，填写 CorpID / Secret / AgentID。"
              />
              <div style="display:flex;align-items:center;gap:10px;flex-wrap:wrap;">
                <el-button type="primary" :loading="bindLoading" @click="buildWeixinBindQr">生成二维码</el-button>
                <el-button @click="load">扫码后刷新列表</el-button>
                <span v-if="form.config.OpengId" class="muted">已绑定用户：{{ form.config.OpengId }}</span>
              </div>
              <div v-if="bindQr" class="qr-box" style="width:240px;min-height:240px;">
                <img :src="bindQr" alt="企业微信扫码绑定二维码" style="width:220px;height:220px;" />
              </div>
              <span v-if="bindHint" class="muted">{{ bindHint }}</span>
            </div>
          </el-form-item>
        </template>
        <el-form-item
          v-for="field in selectedInputs"
          :key="field.key"
          :label="field.name"
        >
          <el-select v-if="field.key === 'ImageMode'" v-model="form.config[field.key]">
            <el-option label="图文卡片" value="news" />
            <el-option label="上传后发送" value="upload" />
            <el-option label="自动" value="auto" />
          </el-select>
          <el-input v-else v-model="form.config[field.key]" :placeholder="field.placeholder" />
        </el-form-item>
      </el-form>
      <div v-if="testResult" class="test-result" :class="testResult.success ? 'ok' : 'fail'">
        <strong>{{ testResult.success ? '测试成功' : '测试失败' }}</strong>
        <span>{{ testResult.message }}</span>
        <span v-if="testResult.statusCode">HTTP {{ testResult.statusCode }}</span>
        <pre v-if="testResult.response">{{ testResult.response }}</pre>
      </div>
      <template #footer>
        <el-button :loading="testing" @click="testCurrent">测试</el-button>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import QRCode from 'qrcode'

import {
  activeSendAuth,
  addSendAuth,
  deleteSendAuth,
  getSendAuths,
  getSendTemplates,
  getWeixinBindUrl,
  modifySendAuth,
  reSendKey,
  testSendAuth
} from '@/api/setting'

const WEIXIN_SCAN_TEMPLATE_ID = 'B1E7D9D4-2A9C-4B5A-8E53-65CC6D8C1F20'

const auths = ref([])
const templates = ref([])
const dialogVisible = ref(false)
const saving = ref(false)
const testing = ref(false)
const bindLoading = ref(false)
const bindQr = ref('')
const bindHint = ref('')
const editing = ref(null)
const form = reactive({ id: 0, templateID: '', name: '', config: {} })
const testResult = ref(null)
const proxyOptions = [
  { label: '不使用', value: 'no' },
  { label: '全局代理', value: 'global' },
  { label: '自定义', value: 'custom' }
]

const selectedTemplate = computed(() => templates.value.find((item) => item.key === form.templateID))
const selectedInputs = computed(() => selectedTemplate.value?.inputs || [])
const isWeixinScanTemplate = computed(() => form.templateID === WEIXIN_SCAN_TEMPLATE_ID)

onMounted(load)

async function load() {
  const [tpls, rows] = await Promise.all([getSendTemplates(), getSendAuths()])
  templates.value = tpls || []
  auths.value = rows || []
}

function openCreate() {
  editing.value = null
  form.id = 0
  form.templateID = templates.value[0]?.key || ''
  form.name = templates.value[0]?.name || ''
  form.config = {}
  bindQr.value = ''
  bindHint.value = ''
  testResult.value = null
  syncFields()
  dialogVisible.value = true
}

function openEdit(row) {
  editing.value = row
  form.id = row.id
  form.templateID = row.templateID
  form.name = row.name
  try {
    form.config = JSON.parse(row.config || '{}')
  } catch {
    form.config = {}
  }
  bindQr.value = ''
  bindHint.value = row.templateID === WEIXIN_SCAN_TEMPLATE_ID && row.config
    ? '若已扫码成功，请刷新列表查看绑定后的企业微信通道。'
    : ''
  testResult.value = null
  syncFields()
  dialogVisible.value = true
}

function handleTemplateChange() {
  if (selectedTemplate.value) form.name = selectedTemplate.value.name
  bindQr.value = ''
  bindHint.value = ''
  testResult.value = null
  syncFields()
}

function syncFields() {
  if (!form.config.ProxyMode) {
    form.config.ProxyMode = form.config.UseProxy ? 'global' : 'no'
  }
  form.config.UseProxy = form.config.ProxyMode !== 'no'
  if (typeof form.config.ProxyAddress === 'undefined') form.config.ProxyAddress = ''
  for (const field of selectedInputs.value) {
    if (typeof form.config[field.key] === 'undefined') form.config[field.key] = ''
  }
  if (typeof form.config.ImageMode !== 'undefined' && form.config.ImageMode === '') {
    form.config.ImageMode = 'news'
  }
}

function buildPayload() {
  return {
    id: form.id,
    sendAuthId: form.id,
    templateID: form.templateID,
    name: form.name,
    config: form.config,
    active: editing.value?.active ?? true
  }
}

async function save() {
  saving.value = true
  try {
    form.config.UseProxy = form.config.ProxyMode !== 'no'
    const payload = buildPayload()
    if (editing.value || form.id > 0) await modifySendAuth(payload)
    else await addSendAuth(payload)
    ElMessage.success('已保存')
    dialogVisible.value = false
    await load()
  } finally {
    saving.value = false
  }
}

async function testCurrent() {
  testing.value = true
  testResult.value = null
  try {
    form.config.UseProxy = form.config.ProxyMode !== 'no'
    const result = await testSendAuth(buildPayload())
    testResult.value = result
    if (result?.success) {
      ElMessage.success('测试消息已发送')
    } else {
      ElMessage.error(result?.message || '测试发送失败，请检查通道配置')
    }
  } finally {
    testing.value = false
  }
}

async function ensureWeixinScanAuthExists() {
  if (form.id > 0) {
    return form.id
  }
  const payload = {
    templateID: WEIXIN_SCAN_TEMPLATE_ID,
    name: form.name || '企业微信扫码绑定',
    config: {},
    active: true
  }
  const created = await addSendAuth(payload)
  form.id = created.id
  editing.value = created
  await load()
  return created.id
}

async function buildWeixinBindQr() {
  bindLoading.value = true
  bindQr.value = ''
  bindHint.value = ''
  try {
    if (!isWeixinScanTemplate.value) {
      ElMessage.warning('请先选择“企业微信扫码绑定”模板')
      return
    }
    const sendAuthId = await ensureWeixinScanAuthExists()
    const bindUrl = await getWeixinBindUrl(sendAuthId, window.location.origin)
    bindQr.value = await QRCode.toDataURL(bindUrl, { width: 220, margin: 1 })
    bindHint.value = '请使用企业微信扫码并授权。授权成功后，页面会显示成功提示，随后返回本页点击“刷新列表”。'
  } catch (error) {
    const msg = error?.message || '生成二维码失败'
    bindHint.value = msg
  } finally {
    bindLoading.value = false
  }
}

async function toggle(row, state) {
  await activeSendAuth(row.id, state)
  ElMessage.success(state ? '已启用' : '已停用')
}

async function resetKey(row) {
  const key = await reSendKey(row.id)
  row.key = key
  ElMessage.success('Key 已重置')
}

async function remove(row) {
  await deleteSendAuth(row.id)
  await load()
  ElMessage.success('已删除')
}

function templateName(id) {
  return templates.value.find((item) => item.key === id)?.name || id
}
</script>
