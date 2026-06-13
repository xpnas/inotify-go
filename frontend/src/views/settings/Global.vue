<template>
  <el-card shadow="never" class="form-card">
    <template #header>全局参数</template>
    <el-form label-width="160px">
      <el-form-item label="GitHub Client ID">
        <el-input v-model="form.githubClientId" />
      </el-form-item>
      <el-form-item label="GitHub Client Secret">
        <el-input v-model="form.githubClientSecret" show-password />
      </el-form-item>
      <el-form-item label="代理地址">
        <el-input v-model="form.proxyAddress" placeholder="http://127.0.0.1:7890" />
      </el-form-item>
      <el-form-item label="管理员账号">
        <el-input v-model="form.administrators" placeholder="admin,githubUserName" />
      </el-form-item>
      <el-form-item label="管理员用户名">
        <el-input v-model="form.adminUserName" placeholder="githubUserName" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'

import { getGlobal, setGlobal } from '@/api/system'

const saving = ref(false)
const form = reactive({ githubClientId: '', githubClientSecret: '', proxyAddress: '', administrators: '', adminUserName: '' })

onMounted(async () => {
  Object.assign(form, await getGlobal())
})

async function save() {
  saving.value = true
  try {
    await setGlobal(form)
    ElMessage.success('全局参数已保存')
  } finally {
    saving.value = false
  }
}
</script>
