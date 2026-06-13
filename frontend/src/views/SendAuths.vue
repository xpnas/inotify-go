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
        <el-form-item
          v-for="field in selectedInputs"
          :key="field.key"
          :label="field.name"
        >
          <el-input v-model="form.config[field.key]" :placeholder="field.placeholder" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'

import {
  activeSendAuth,
  addSendAuth,
  deleteSendAuth,
  getSendAuths,
  getSendTemplates,
  modifySendAuth,
  reSendKey
} from '@/api/setting'

const auths = ref([])
const templates = ref([])
const dialogVisible = ref(false)
const saving = ref(false)
const editing = ref(null)
const form = reactive({ id: 0, templateID: '', name: '', config: {} })

const selectedTemplate = computed(() => templates.value.find((item) => item.key === form.templateID))
const selectedInputs = computed(() => selectedTemplate.value?.inputs || [])

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
  syncFields()
  dialogVisible.value = true
}

function handleTemplateChange() {
  if (selectedTemplate.value) form.name = selectedTemplate.value.name
  syncFields()
}

function syncFields() {
  for (const field of selectedInputs.value) {
    if (typeof form.config[field.key] === 'undefined') form.config[field.key] = ''
  }
}

async function save() {
  saving.value = true
  try {
    const payload = {
      id: form.id,
      sendAuthId: form.id,
      templateID: form.templateID,
      name: form.name,
      config: form.config,
      active: editing.value?.active ?? true
    }
    if (editing.value) await modifySendAuth(payload)
    else await addSendAuth(payload)
    ElMessage.success('已保存')
    dialogVisible.value = false
    await load()
  } finally {
    saving.value = false
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
