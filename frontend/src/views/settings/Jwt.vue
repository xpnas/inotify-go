<template>
  <el-card shadow="never" class="form-card">
    <template #header>JWT 参数</template>
    <el-form label-width="160px">
      <el-form-item label="Issuer">
        <el-input v-model="form.issuer" />
      </el-form-item>
      <el-form-item label="Audience">
        <el-input v-model="form.audience" />
      </el-form-item>
      <el-form-item label="IssuerSigningKey">
        <el-input v-model="form.issuerSigningKey" show-password />
      </el-form-item>
      <el-form-item label="过期时间(分钟)">
        <el-input-number v-model="form.accessTokenExpires" :min="1" :step="60" />
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

import { getJWT, setJWT } from '@/api/system'

const saving = ref(false)
const form = reactive({ issuer: '', audience: '', issuerSigningKey: '', accessTokenExpires: 43200 })

onMounted(async () => {
  Object.assign(form, await getJWT())
})

async function save() {
  saving.value = true
  try {
    await setJWT(form)
    ElMessage.success('JWT 参数已保存')
  } finally {
    saving.value = false
  }
}
</script>
