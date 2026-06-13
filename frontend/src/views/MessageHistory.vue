<template>
  <div class="stack">
    <el-card shadow="never">
      <el-form class="history-filter" :model="filters" label-width="72px">
        <el-form-item label="标题">
          <el-input v-model="filters.title" clearable placeholder="标题关键字" @keyup.enter="search" />
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="filters.content" clearable placeholder="标题或内容关键字" @keyup.enter="search" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.success" clearable placeholder="全部">
            <el-option label="成功" value="true" />
            <el-option label="失败" value="false" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
          />
        </el-form-item>
        <el-form-item class="history-actions">
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="reset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <el-table v-loading="loading" :data="rows" border>
        <el-table-column prop="createTime" label="发送时间" width="180">
          <template #default="{ row }">{{ formatTime(row.createTime) }}</template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip />
        <el-table-column prop="body" label="内容" min-width="260">
          <template #default="{ row }">
            <div class="history-body">{{ row.body }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="url" label="URL" min-width="180" show-overflow-tooltip />
        <el-table-column prop="group" label="分组" min-width="120" show-overflow-tooltip />
        <el-table-column prop="channelCount" label="通道数" width="92" align="center" />
        <el-table-column prop="success" label="状态" width="92" align="center">
          <template #default="{ row }">
            <el-tag :type="row.success ? 'success' : 'danger'" effect="plain">
              {{ row.success ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-row">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @size-change="load"
          @current-change="load"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'

import { getMessageHistories } from '@/api/setting'

const loading = ref(false)
const rows = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const dateRange = ref([])
const filters = reactive({
  title: '',
  content: '',
  success: ''
})

onMounted(load)

async function load() {
  loading.value = true
  try {
    const [startTime, endTime] = dateRange.value || []
    const data = await getMessageHistories({
      page: page.value,
      pageSize: pageSize.value,
      title: filters.title || undefined,
      content: filters.content || undefined,
      success: filters.success || undefined,
      startTime,
      endTime
    })
    rows.value = data.items || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  load()
}

function reset() {
  filters.title = ''
  filters.content = ''
  filters.success = ''
  dateRange.value = []
  search()
}

function formatTime(value) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}
</script>
