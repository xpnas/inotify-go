<template>
  <div class="stack">
    <div class="metric-grid">
      <el-card shadow="never">
        <span>今日发送</span>
        <strong>{{ todayCount }}</strong>
      </el-card>
      <el-card shadow="never">
        <span>累计发送</span>
        <strong>{{ totalCount }}</strong>
      </el-card>
      <el-card shadow="never">
        <span>通道类型</span>
        <strong>{{ types.length }}</strong>
      </el-card>
    </div>
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>发送统计</span>
          <el-button @click="load">刷新</el-button>
        </div>
      </template>
      <el-table :data="infos" border>
        <el-table-column prop="date" label="日期" width="160" />
        <el-table-column prop="templateID" label="模板" min-width="220">
          <template #default="{ row }">{{ typeName(row.templateID) }}</template>
        </el-table-column>
        <el-table-column prop="count" label="次数" width="120" />
      </el-table>
    </el-card>

    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>系统诊断</span>
          <div>
            <el-button @click="downloadBackup">备份数据库</el-button>
            <el-button @click="load">刷新</el-button>
          </div>
        </div>
      </template>
      <el-alert
        v-if="diagnosticsError"
        type="warning"
        :closable="false"
        title="诊断接口暂不可用"
        description="请确认后端服务已更新并重启；发送统计仍可正常查看。"
      />
      <el-descriptions v-else-if="diagnostics" :column="1" border>
        <el-descriptions-item label="GitHub OAuth">
          <el-tag :type="diagnostics.githubConfigured ? 'success' : 'warning'" effect="plain">
            {{ diagnostics.githubConfigured ? '已配置' : '未配置' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="企业微信登录">
          <el-tag :type="diagnostics.weixinConfigured ? 'success' : 'warning'" effect="plain">
            {{ diagnostics.weixinConfigured ? '已配置' : '未配置' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="GitHub 回调">{{ diagnostics.githubCallback }}</el-descriptions-item>
        <el-descriptions-item label="企业微信回调">{{ diagnostics.weixinCallback }}</el-descriptions-item>
        <el-descriptions-item label="代理">
          <el-tag :type="diagnostics.proxyValid ? 'success' : 'danger'" effect="plain">
            {{ diagnostics.proxyConfigured ? (diagnostics.proxyValid ? '格式有效' : '格式无效') : '未配置' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="数据目录">
          {{ diagnostics.dataDir }}
          <el-tag :type="diagnostics.dataDirWritable ? 'success' : 'danger'" effect="plain">
            {{ diagnostics.dataDirWritable ? '可写' : '不可写' }}
          </el-tag>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'

import { backupDatabase, getDiagnostics, getSendInfos, getSendTypeInfos } from '@/api/system'

const infos = ref([])
const types = ref([])
const diagnostics = ref(null)
const diagnosticsError = ref(false)
const today = new Date().toISOString().slice(0, 10)
const totalCount = computed(() => infos.value.reduce((sum, item) => sum + Number(item.count || 0), 0))
const todayCount = computed(() => infos.value.filter((item) => item.date === today).reduce((sum, item) => sum + Number(item.count || 0), 0))

onMounted(load)

async function load() {
  const [rows, typeRows] = await Promise.all([getSendInfos(), getSendTypeInfos()])
  infos.value = rows || []
  types.value = typeRows || []
  diagnosticsError.value = false
  try {
    diagnostics.value = await getDiagnostics()
  } catch {
    diagnostics.value = null
    diagnosticsError.value = true
  }
}

function typeName(id) {
  return types.value.find((item) => item.key === id)?.name || id
}

async function downloadBackup() {
  const blob = await backupDatabase()
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `inotify-backup-${today}.db`
  link.click()
  URL.revokeObjectURL(url)
}
</script>
