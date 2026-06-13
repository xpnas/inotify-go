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
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'

import { getSendInfos, getSendTypeInfos } from '@/api/system'

const infos = ref([])
const types = ref([])
const today = new Date().toISOString().slice(0, 10)
const totalCount = computed(() => infos.value.reduce((sum, item) => sum + Number(item.count || 0), 0))
const todayCount = computed(() => infos.value.filter((item) => item.date === today).reduce((sum, item) => sum + Number(item.count || 0), 0))

onMounted(load)

async function load() {
  const [rows, typeRows] = await Promise.all([getSendInfos(), getSendTypeInfos()])
  infos.value = rows || []
  types.value = typeRows || []
}

function typeName(id) {
  return types.value.find((item) => item.key === id)?.name || id
}
</script>
