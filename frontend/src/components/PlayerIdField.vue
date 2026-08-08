<template>
  <div class="player-id-field">
    <el-autocomplete
      :model-value="modelValue"
      :fetch-suggestions="queryPlayerIds"
      :placeholder="placeholder"
      class="player-id-input"
      clearable
      :trigger-on-focus="true"
      value-key="value"
      @update:model-value="handleInput"
    >
      <template #append>
        <el-tooltip content="选择未停用账号" placement="top">
          <el-button
            :icon="User"
            :loading="accountsLoading && dialogVisible"
            aria-label="选择未停用账号"
            @click.stop="openDialog"
          />
        </el-tooltip>
      </template>
    </el-autocomplete>
  </div>

  <el-dialog
    v-model="dialogVisible"
    title="选择账号"
    width="min(720px, 92vw)"
    top="4vh"
    append-to-body
    class="account-picker-dialog"
  >
    <div class="account-picker-toolbar">
      <el-input
        v-model="accountQuery"
        :prefix-icon="Search"
        placeholder="搜索缩写 / ID / 尾号 / 昵称"
        clearable
      />
      <el-button :icon="Refresh" :loading="accountsLoading" @click="loadAccounts(true)">
        刷新
      </el-button>
    </div>

    <el-alert
      v-if="accountsError"
      :title="accountsError"
      type="warning"
      show-icon
      :closable="false"
      class="account-picker-error"
    />

    <el-table
      v-loading="accountsLoading"
      :data="filteredAccounts"
      max-height="min(680px, calc(100vh - 220px))"
      empty-text="暂无未停用账号"
      stripe
      @row-click="selectAccount"
    >
      <el-table-column label="" width="88" fixed="left">
        <template #default="{ row }">
          <el-button type="primary" size="small" @click.stop="selectAccount(row)">
            选择
          </el-button>
        </template>
      </el-table-column>
      <el-table-column prop="abbr" label="缩写" width="110">
        <template #default="{ row }">
          <strong>{{ row.abbr || '-' }}</strong>
        </template>
      </el-table-column>
      <el-table-column prop="phone_tail" label="尾号" width="90">
        <template #default="{ row }">
          {{ row.phone_tail || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="体力" width="110">
        <template #default="{ row }">
          <span
            class="energy-pill"
            :class="getEnergyStageClass(row)"
            :title="getEnergyTooltip(row)"
          >
            {{ formatEnergy(row) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column prop="id" label="玩家ID" width="150" />
      <el-table-column prop="nickname" label="游戏内昵称" min-width="150">
        <template #default="{ row }">
          {{ row.nickname || '-' }}
        </template>
      </el-table-column>
    </el-table>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Refresh, Search, User } from '@element-plus/icons-vue'
import { accountsApi, type ActiveAccount } from '../api'

const props = withDefaults(defineProps<{
  modelValue?: string
  playerIds?: string[]
  placeholder?: string
}>(), {
  modelValue: '',
  playerIds: () => [],
  placeholder: '例如: 120003177'
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const dialogVisible = ref(false)
const accounts = ref<ActiveAccount[]>([])
const accountsLoaded = ref(false)
const accountsLoading = ref(false)
const accountsError = ref('')
const accountQuery = ref('')

const normalize = (value: unknown) => String(value ?? '').trim().toLowerCase()

const accountIdsForSuggestions = computed(() => {
  const ids = [
    ...props.playerIds,
    ...accounts.value.map((account) => account.id)
  ].map((id) => id.trim()).filter(Boolean)

  return Array.from(new Set(ids))
})

const filteredAccounts = computed(() => {
  const query = normalize(accountQuery.value)
  if (!query) {
    return accounts.value
  }

  return accounts.value.filter((account) => {
    const searchText = [
      account.abbr,
      account.id,
      account.phone_tail,
      account.nickname
    ].map(normalize).join(' ')

    return searchText.includes(query)
  })
})

const handleInput = (value: string) => {
  emit('update:modelValue', value)
}

const queryPlayerIds = (queryString: string, cb: (results: { value: string }[]) => void) => {
  const query = normalize(queryString)
  const results = accountIdsForSuggestions.value
    .filter((id) => normalize(id).includes(query))
    .slice(0, 10)
    .map((id) => ({ value: id }))

  if (queryString && results.length === 0) {
    results.push({ value: queryString })
  }

  cb(results)
}

const getErrorMessage = (error: unknown) => {
  const response = (error as { response?: { data?: { detail?: string } } })?.response
  return response?.data?.detail || (error instanceof Error ? error.message : '加载账号失败')
}

const formatEnergy = (account: ActiveAccount) => {
  const waveplate = Number(account.current_waveplate || 0)
  const crystal = Number(account.current_waveplate_crystal || 0)
  return crystal > 0 ? `${waveplate} + ${crystal}` : String(waveplate)
}

const getEnergyTotal = (account: ActiveAccount) => {
  const waveplate = Number(account.current_waveplate || 0)
  const crystal = Number(account.current_waveplate_crystal || 0)
  return waveplate + crystal
}

const getEnergyStageClass = (account: ActiveAccount) => {
  const total = getEnergyTotal(account)
  if (total >= 240) return 'energy-stage-high'
  if (total >= 120) return 'energy-stage-mid'
  return 'energy-stage-low'
}

const getEnergyTooltip = (account: ActiveAccount) => {
  const waveplate = Number(account.current_waveplate || 0)
  const crystal = Number(account.current_waveplate_crystal || 0)
  const total = waveplate + crystal
  return `当前体力 ${waveplate}，体力结晶 ${crystal}，总计 ${total}`
}

const loadAccounts = async (force = false) => {
  if (accountsLoaded.value && !force) {
    return
  }

  accountsLoading.value = true
  accountsError.value = ''
  try {
    const response = await accountsApi.getActiveAccounts()
    accounts.value = response.data
    accountsLoaded.value = true
  } catch (error) {
    accountsError.value = getErrorMessage(error)
  } finally {
    accountsLoading.value = false
  }
}

const openDialog = () => {
  dialogVisible.value = true
  void loadAccounts()
}

const selectAccount = (account: ActiveAccount) => {
  emit('update:modelValue', account.id)
  dialogVisible.value = false
}

defineExpose({
  openDialog
})
</script>

<style scoped>
.player-id-field {
  width: 100%;
}

.player-id-input {
  width: 100%;
}

:deep(.el-input-group__append) {
  padding: 0;
}

:deep(.el-input-group__append .el-button) {
  margin: 0;
  min-height: 30px;
  padding: 0 12px;
}

.account-picker-toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.account-picker-error {
  margin-bottom: 12px;
}

.energy-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 68px;
  padding: 2px 8px;
  border: 1px solid transparent;
  border-radius: 999px;
  font-weight: 700;
  line-height: 1.2;
  white-space: nowrap;
}

.energy-stage-low {
  background: #eef6ff;
  border-color: #bfdbfe;
  color: #1d4ed8;
}

.energy-stage-mid {
  background: #fff7ed;
  border-color: #fdba74;
  color: #9a3412;
}

.energy-stage-high {
  background: #fef2f2;
  border-color: #f87171;
  color: #b91c1c;
}

:deep(.el-table__row) {
  cursor: pointer;
}

@media (max-width: 640px) {
  .account-picker-toolbar {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
