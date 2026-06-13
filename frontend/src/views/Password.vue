<template>
  <el-card shadow="never" class="form-card">
    <template #header>重置密码</template>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item label="用户名">
        <el-input v-model="form.username" disabled />
      </el-form-item>
      <el-form-item label="新密码" prop="password">
        <el-input v-model="form.password" type="password" show-password />
      </el-form-item>
      <el-form-item label="确认密码" prop="confirm">
        <el-input v-model="form.confirm" type="password" show-password />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="loading" @click="submit">保存</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'

import { resetPassword } from '@/api/user'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const formRef = ref()
const loading = ref(false)
const form = reactive({ username: auth.name, password: '', confirm: '' })
const rules = {
  password: [{ required: true, min: 6, message: '至少 6 位', trigger: 'blur' }],
  confirm: [
    {
      validator: (_, value, callback) => {
        if (value !== form.password) callback(new Error('两次输入不一致'))
        else callback()
      },
      trigger: 'blur'
    }
  ]
}

async function submit() {
  await formRef.value.validate()
  loading.value = true
  try {
    await resetPassword({ username: form.username, password: form.password })
    ElMessage.success('密码已更新')
    form.password = ''
    form.confirm = ''
  } finally {
    loading.value = false
  }
}
</script>
