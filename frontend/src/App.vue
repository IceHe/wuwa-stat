<template>
  <el-container class="container">
    <el-header>
      <div class="header-content">
        <h1>鸣潮产出统计</h1>
        <el-button v-if="isLoggedIn" @click="handleLogout">退出登录</el-button>
      </div>
    </el-header>

    <el-main>
      <el-card v-if="!isLoggedIn" class="login-card">
        <template #header>
          <span>Token 登录</span>
        </template>
        <el-form @submit.prevent>
          <el-form-item label="Token">
            <el-input
              v-model="tokenInput"
              type="password"
              show-password
              placeholder="请输入 token"
              @keyup.enter="handleLogin"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="authLoading" @click="handleLogin">登录</el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <el-result
        v-else-if="!canView"
        icon="warning"
        title="无查看权限"
        sub-title="当前账号没有 view 权限"
      />

      <el-tabs v-else v-model="activeTab">
        <el-tab-pane label="无音区产出统计" name="tacet" lazy>
          <TacetRecordInput v-if="canEdit" @success="handleInputSuccess" />
          <TacetRecordList :refresh="refreshTrigger" :can-edit="canEdit" class="mt-16" />
          <TacetStatsView :refresh="refreshTrigger" class="mt-16" />
        </el-tab-pane>

        <el-tab-pane label="共鸣者突破材料统计" name="ascension" lazy>
          <AscensionRecordInput v-if="canEdit" @success="handleAscensionInputSuccess" />
          <AscensionRecordList :refresh="ascensionRefreshTrigger" :can-edit="canEdit" class="mt-16" />
          <AscensionStatsView :refresh="ascensionRefreshTrigger" class="mt-16" />
        </el-tab-pane>

        <el-tab-pane label="凝素领域产出统计" name="resonance" lazy>
          <ResonanceRecordInput v-if="canEdit" @success="handleResonanceInputSuccess" />
          <ResonanceRecordList :refresh="resonanceRefreshTrigger" :can-edit="canEdit" class="mt-16" />
          <ResonanceStatsView :refresh="resonanceRefreshTrigger" class="mt-16" />
        </el-tab-pane>
      </el-tabs>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import TacetRecordInput from './components/TacetRecordInput.vue'
import TacetRecordList from './components/TacetRecordList.vue'
import TacetStatsView from './components/TacetStatsView.vue'
import AscensionRecordInput from './components/AscensionRecordInput.vue'
import AscensionRecordList from './components/AscensionRecordList.vue'
import AscensionStatsView from './components/AscensionStatsView.vue'
import ResonanceRecordInput from './components/ResonanceRecordInput.vue'
import ResonanceRecordList from './components/ResonanceRecordList.vue'
import ResonanceStatsView from './components/ResonanceStatsView.vue'
import {
  authApi,
  authEvents,
  clearStoredAuthToken,
  getStoredAuthToken,
  setStoredAuthToken,
  type Permission
} from './api'

const activeTab = ref('tacet')
const refreshTrigger = ref(0)
const ascensionRefreshTrigger = ref(0)
const resonanceRefreshTrigger = ref(0)
const tokenInput = ref('')
const authLoading = ref(false)
const isLoggedIn = ref(false)
const permissions = ref<Permission[]>([])

const canView = computed(() => permissions.value.includes('view') || permissions.value.includes('manage'))
const canEdit = computed(() => permissions.value.includes('edit') || permissions.value.includes('manage'))

const handleInputSuccess = () => {
  refreshTrigger.value++
}

const handleAscensionInputSuccess = () => {
  ascensionRefreshTrigger.value++
}

const handleResonanceInputSuccess = () => {
  resonanceRefreshTrigger.value++
}

const resetAuthState = () => {
  isLoggedIn.value = false
  permissions.value = []
  tokenInput.value = ''
}

const handleLogout = () => {
  clearStoredAuthToken()
  resetAuthState()
}

const restoreSession = async () => {
  const token = getStoredAuthToken()
  if (!token) {
    return
  }

  authLoading.value = true
  try {
    tokenInput.value = token
    const authPermissions = await authApi.me()
    permissions.value = authPermissions
    isLoggedIn.value = true
  } catch {
    clearStoredAuthToken()
    resetAuthState()
  } finally {
    authLoading.value = false
  }
}

const handleLogin = async () => {
  const token = tokenInput.value.trim()
  if (!token) {
    ElMessage.warning('请输入 token')
    return
  }

  authLoading.value = true
  try {
    setStoredAuthToken(token)
    const authPermissions = await authApi.me()
    permissions.value = authPermissions
    isLoggedIn.value = true
    ElMessage.success('登录成功')
  } catch {
    clearStoredAuthToken()
    resetAuthState()
    ElMessage.error('token 无效、已过期或鉴权服务不可用')
  } finally {
    authLoading.value = false
  }
}

const handleUnauthorized = () => {
  const wasLoggedIn = isLoggedIn.value
  resetAuthState()
  if (wasLoggedIn) {
    ElMessage.warning('登录已失效，请重新登录')
  }
}

onMounted(async () => {
  window.addEventListener(authEvents.unauthorized, handleUnauthorized)
  await restoreSession()
})

onBeforeUnmount(() => {
  window.removeEventListener(authEvents.unauthorized, handleUnauthorized)
})
</script>

<style scoped>
.container {
  min-height: 100vh;
  background: #f5f7fa;
}

.el-header {
  background: #409eff;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
}

.header-content {
  width: min(1200px, 95vw);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

h1 {
  margin: 0;
  font-size: 20px;
}

.login-card {
  max-width: 560px;
  margin: 0 auto;
}

.mt-16 {
  margin-top: 12px;
}
</style>
