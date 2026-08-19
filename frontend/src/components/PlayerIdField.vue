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
      @focus="refreshAccount"
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

    <dl v-if="matchedAccount" class="account-summary" aria-live="polite">
      <div class="account-summary-item">
        <dt>缩写</dt>
        <dd>
          <strong>{{ matchedAccount.abbr || '-' }}</strong>
          <span v-if="!matchedAccount.is_active" class="inactive-label">已停用</span>
        </dd>
      </div>
      <div class="account-summary-item">
        <dt>尾号</dt>
        <dd>{{ matchedAccount.phone_tail || '-' }}</dd>
      </div>
      <div class="account-summary-item">
        <dt>体力</dt>
        <dd>
          <span
            class="energy-pill"
            :class="getEnergyStageClass(matchedAccount)"
            :title="getEnergyTooltip(matchedAccount)"
          >
            {{ formatEnergy(matchedAccount) }}
          </span>
        </dd>
      </div>
      <div class="account-summary-item">
        <dt>玩家ID</dt>
        <dd>{{ matchedAccount.id }}</dd>
      </div>
      <div class="account-summary-item account-summary-nickname">
        <dt>游戏内昵称</dt>
        <dd>{{ matchedAccount.nickname || '-' }}</dd>
      </div>
    </dl>
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
      <el-button :icon="Refresh" :loading="accountsLoading" @click="loadAccounts">
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
import { computed, onBeforeUnmount, ref, watch } from 'vue'
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
const accountsLoading = ref(false)
const accountsError = ref('')
const accountQuery = ref('')
const matchedAccount = ref<ActiveAccount | null>(null)
let accountLookupTimer: ReturnType<typeof setTimeout> | null = null
let accountLookupSequence = 0

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

const getResponseStatus = (error: unknown) =>
  (error as { response?: { status?: number } })?.response?.status

const lookupAccount = async (playerId: string, sequence: number) => {
  try {
    const response = await accountsApi.getAccountById(playerId)
    if (sequence === accountLookupSequence && normalize(props.modelValue) === normalize(playerId)) {
      matchedAccount.value = response.data
    }
  } catch (error) {
    if (sequence === accountLookupSequence && getResponseStatus(error) === 404) {
      matchedAccount.value = null
    }
  }
}

const scheduleAccountLookup = (value: string) => {
  if (accountLookupTimer) {
    clearTimeout(accountLookupTimer)
    accountLookupTimer = null
  }

  const playerId = String(value ?? '').trim()
  const sequence = ++accountLookupSequence
  matchedAccount.value = accounts.value.find((account) => normalize(account.id) === normalize(playerId)) || null
  if (!playerId) {
    return
  }

  accountLookupTimer = setTimeout(() => {
    accountLookupTimer = null
    void lookupAccount(playerId, sequence)
  }, 400)
}

const refreshAccount = async () => {
  if (accountLookupTimer) {
    clearTimeout(accountLookupTimer)
    accountLookupTimer = null
  }

  const playerId = String(props.modelValue ?? '').trim()
  const sequence = ++accountLookupSequence
  if (!playerId) {
    matchedAccount.value = null
    return
  }

  await lookupAccount(playerId, sequence)
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

const loadAccounts = async () => {
  accountsLoading.value = true
  accountsError.value = ''
  try {
    const response = await accountsApi.getActiveAccounts()
    accounts.value = response.data
    const currentPlayerId = normalize(props.modelValue)
    const currentAccount = response.data.find((account) => normalize(account.id) === currentPlayerId)
    if (currentAccount) {
      matchedAccount.value = currentAccount
    }
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
  matchedAccount.value = account
  emit('update:modelValue', account.id)
  dialogVisible.value = false
}

watch(
  () => props.modelValue,
  scheduleAccountLookup,
  { immediate: true }
)

onBeforeUnmount(() => {
  accountLookupSequence++
  if (accountLookupTimer) {
    clearTimeout(accountLookupTimer)
  }
})

defineExpose({
  openDialog,
  refreshAccount
})
</script>

<style scoped>
.player-id-field {
  width: 100%;
}

.player-id-input {
  width: 100%;
}

.account-summary {
  display: grid;
  grid-template-columns: minmax(64px, 0.65fr) minmax(64px, 0.65fr) minmax(90px, 0.8fr) minmax(110px, 1.1fr) minmax(110px, 1.25fr);
  gap: 8px;
  margin: 8px 0 0;
  padding: 8px 10px;
  background: #f8fafc;
  border: 1px solid #e5e7eb;
  border-left: 3px solid #409eff;
  border-radius: 4px;
}

.account-summary-item {
  min-width: 0;
}

.account-summary-item dt {
  margin: 0 0 2px;
  color: #909399;
  font-size: 12px;
  line-height: 1.3;
}

.account-summary-item dd {
  display: flex;
  align-items: center;
  gap: 6px;
  min-height: 24px;
  margin: 0;
  color: #303133;
  font-size: 13px;
  line-height: 1.4;
  overflow-wrap: anywhere;
}

.inactive-label {
  color: #c45656;
  font-size: 12px;
  white-space: nowrap;
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
  .account-summary {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .account-summary-nickname {
    grid-column: 1 / -1;
  }

  .account-picker-toolbar {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
