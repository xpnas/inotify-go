<template>
  <el-card shadow="never">
    <template #header>
      <div class="card-header">
        <span>用户管理</span>
        <el-button @click="load">刷新</el-button>
      </div>
    </template>
    <el-table :data="users" border>
      <el-table-column prop="userName" label="用户名" min-width="140" />
      <el-table-column prop="email" label="邮箱" min-width="180" />
      <el-table-column prop="githubLogin" label="GitHub" min-width="140">
        <template #default="{ row }">{{ row.githubLogin || '-' }}</template>
      </el-table-column>
      <el-table-column prop="weixinId" label="企业微信" min-width="160">
        <template #default="{ row }">{{ row.weixinId || '-' }}</template>
      </el-table-column>
      <el-table-column prop="token" label="Token" min-width="260" />
      <el-table-column prop="active" label="激活" width="96" align="center">
        <template #default="{ row }">
          <el-switch v-model="row.active" :disabled="row.userName === 'admin'" @change="(value) => toggle(row, value)" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="120">
        <template #default="{ row }">
          <el-popconfirm title="确认删除用户？" @confirm="remove(row)">
            <template #reference>
              <el-button link type="danger" :disabled="row.userName === 'admin'">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'

import { activeUser, deleteUser, getUsers } from '@/api/system'

const users = ref([])

onMounted(load)

async function load() {
  users.value = (await getUsers()) || []
}

async function toggle(row, state) {
  await activeUser(row.userName, state)
  ElMessage.success(state ? '已激活' : '已停用')
}

async function remove(row) {
  await deleteUser(row.userName)
  await load()
  ElMessage.success('已删除')
}
</script>
