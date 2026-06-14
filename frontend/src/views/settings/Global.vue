<template>
  <div class="stack">
    <!-- 界面设置 -->
    <el-card shadow="never" class="form-card">
      <template #header>界面设置</template>
      <el-form label-width="120px">
        <el-form-item label="品牌图标">
          <div class="brand-icon-row">
            <span class="brand-preview">
              <span class="brand-icon">{{ brandForm.icon }}</span>
              {{ brandForm.name }}
            </span>
            <el-input v-model="brandForm.icon" placeholder="输入 Emoji，如 🔔 🚀 📢" style="max-width:200px" />
          </div>
          <div class="form-tip muted">支持任意 Emoji 字符，留空恢复默认 🔔</div>
        </el-form-item>
        <el-form-item label="品牌名称">
          <el-input v-model="brandForm.name" placeholder="Inotify" style="max-width:260px" />
          <div class="form-tip muted">显示在侧边栏左上角，留空恢复默认 Inotify</div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="saveBrand">应用</el-button>
          <el-button @click="resetBrand">还原默认</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- GitHub 登录设置 -->
    <el-card shadow="never" class="form-card">
      <template #header>GitHub 登录设置</template>
      <el-form label-width="160px">
        <el-form-item label="GitHub Client ID">
          <el-input v-model="form.githubClientId" />
        </el-form-item>
        <el-form-item label="GitHub Client Secret">
          <el-input v-model="form.githubClientSecret" show-password />
        </el-form-item>
        <el-form-item label="GitHub redirect_uri">
          <el-input :model-value="githubRedirectUri" readonly />
        </el-form-item>
        <el-form-item label="代理地址">
          <el-input v-model="form.proxyAddress" placeholder="http://127.0.0.1:7890" />
          <div class="form-tip muted">通过代理访问 GitHub OAuth 接口，留空则直连。</div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="saving" @click="save">保存</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 管理员设置 -->
    <el-card shadow="never" class="form-card">
      <template #header>管理员设置</template>
      <el-form label-width="160px">
        <el-form-item label="管理员账号">
          <el-input v-model="form.administrators" placeholder="admin,githubUserName" />
          <div class="form-tip muted">多个账号用英文逗号分隔，拥有系统管理权限。</div>
        </el-form-item>
        <el-form-item label="管理员用户名">
          <el-input v-model="form.adminUserName" placeholder="githubUserName" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="saving" @click="save">保存</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 企业微信扫码绑定配置 -->
    <el-card shadow="never" class="form-card">
      <template #header>企业微信扫码绑定配置</template>
      <el-form label-width="160px">
        <el-form-item label="企业ID CorpID">
          <el-input v-model="form.weixinCorpId" placeholder="wwxxxxxxxxxxxx" />
        </el-form-item>
        <el-form-item label="应用 Secret">
          <el-input v-model="form.weixinCorpSecret" show-password placeholder="企业应用 Secret" />
        </el-form-item>
        <el-form-item label="应用 AgentID">
          <el-input v-model="form.weixinAgentId" placeholder="1000002" />
          <div class="form-tip muted">新增通道时选择“企业微信扫码绑定”模板，将读取此处配置生成二维码。</div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="saving" @click="save">保存</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'

import { getGlobal, setGlobal } from '@/api/system'
import { useBrandStore } from '@/stores/brand'

const saving = ref(false)
const form = reactive({
  githubClientId: '',
  githubClientSecret: '',
  weixinCorpId: '',
  weixinCorpSecret: '',
  weixinAgentId: '',
  proxyAddress: '',
  administrators: '',
  adminUserName: ''
})

const brand = useBrandStore()
const brandForm = reactive({ icon: brand.icon, name: brand.name })
const githubRedirectUri = computed(() => `${window.location.origin}/oauth/github/callback`)

onMounted(async () => {
  Object.assign(form, await getGlobal())
})

async function save() {
  saving.value = true
  try {
    await setGlobal(form)
    ElMessage.success('系统参数已保存')
  } finally {
    saving.value = false
  }
}

function saveBrand() {
  brand.save(brandForm.icon, brandForm.name)
  ElMessage.success('界面设置已应用')
}

function resetBrand() {
  brandForm.icon = '🔔'
  brandForm.name = 'Inotify'
  brand.save('🔔', 'Inotify')
  ElMessage.success('已还原默认')
}
</script>
