<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span>录入产出记录</span>
      </div>
    </template>

    <el-form :model="form" label-width="120px">
      <el-form-item label="日期">
        <el-date-picker
          v-model="form.date"
          type="date"
          placeholder="选择日期"
          value-format="YYYY-MM-DD"
          @change="handleDateChange"
          style="width: 100%"
        />
      </el-form-item>

      <el-form-item label="玩家ID">
        <el-autocomplete
          v-model="form.player_id"
          :fetch-suggestions="queryPlayerIds"
          placeholder="例如: 120003177"
          style="width: 100%"
          clearable
          :trigger-on-focus="true"
          value-key="value"
        />
      </el-form-item>

      <el-form-item label="索拉等级">
        <div class="option-button-group">
          <el-button
            v-for="level in solaLevels"
            :key="level"
            :type="form.sola_level === level ? 'primary' : 'default'"
            @click="form.sola_level !== level && ((form.sola_level = level), handleLevelChange())"
          >
            等级 {{ level }}
          </el-button>
        </div>
      </el-form-item>

      <el-form-item label="领取次数">
        <div class="option-button-group">
          <el-button
            :type="form.claim_count === 1 ? 'primary' : 'default'"
            @click="handleClaimCountChange(1)"
          >
            1次领取
          </el-button>
          <el-button
            :type="form.claim_count === 2 ? 'primary' : 'default'"
            @click="handleClaimCountChange(2)"
          >
            2次领取
          </el-button>
        </div>
      </el-form-item>

      <el-form-item label="掉落组合">
        <div class="option-button-group">
          <el-button
            v-for="combo in availableCombos"
            :key="combo.key"
            :type="selectedComboKey === combo.key ? 'primary' : 'default'"
            @click="selectedComboKey !== combo.key && ((selectedComboKey = combo.key), handleComboChange())"
          >
            <span class="material-gold">金{{ combo.gold }}</span>
            <span> </span>
            <span class="material-purple">紫{{ combo.purple }}</span>
          </el-button>
        </div>
        <div class="combo-hint" v-if="currentCombo">
          <span class="material-gold">金{{ currentCombo.gold }}</span>
          <span> </span>
          <span class="material-purple">紫{{ currentCombo.purple }}</span>
        </div>
        <div class="exp-hint">
          {{ form.claim_count === 1 ? '单次领取组合' : '两次领取合并后的组合' }}
        </div>
        <div class="exp-hint" v-if="currentCombo">
          声骸经验：{{ currentCombo.experience.toLocaleString() }}
        </div>
      </el-form-item>

      <el-form-item label="金色密音筒" class="material-item-gold">
        <el-input-number v-model="form.gold_tubes" :min="0" disabled />
      </el-form-item>

      <el-form-item label="紫色密音筒" class="material-item-purple">
        <el-input-number v-model="form.purple_tubes" :min="0" disabled />
      </el-form-item>

      <el-form-item>
        <el-button type="primary" @click="handleSubmit" :loading="loading">
          提交
        </el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import { recordApi } from '../api'

const emit = defineEmits(['success'])

const STORAGE_KEY = 'wuwa_last_player_id'
const playerIds = ref<string[]>([])

// 从localStorage获取上次使用的玩家ID
const getStoredPlayerId = (): string | null => {
  return localStorage.getItem(STORAGE_KEY)
}

// 保存玩家ID到localStorage
const savePlayerId = (playerId: string) => {
  localStorage.setItem(STORAGE_KEY, playerId)
}

// 固定掉落组合表（根据提供的表格）
const combosByLevel: Record<number, { gold: number; purple: number; experience: number }[]> = {
  8: [
    { gold: 4, purple: 4, experience: 28000 },
    { gold: 3, purple: 4, experience: 23000 }
  ],
  7: [
    { gold: 4, purple: 4, experience: 28000 },
    { gold: 4, purple: 3, experience: 26000 },
    { gold: 3, purple: 4, experience: 23000 },
    { gold: 3, purple: 3, experience: 21000 }
  ],
  6: [
    { gold: 4, purple: 4, experience: 28000 },
    { gold: 4, purple: 3, experience: 26000 },
    { gold: 3, purple: 4, experience: 23000 },
    { gold: 3, purple: 3, experience: 21000 }
  ],
  5: [
    { gold: 3, purple: 6, experience: 27000 },
    { gold: 3, purple: 5, experience: 25000 },
    { gold: 2, purple: 6, experience: 22000 },
    { gold: 2, purple: 5, experience: 20000 }
  ]
}

const solaLevels = [8, 7, 6, 5]

const loading = ref(false)
const isDateManuallyEdited = ref(false)
let gameDateTimer: ReturnType<typeof setTimeout> | null = null

const formatLocalDate = (date: Date) => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const getDefaultGameDate = () => {
  const now = new Date()
  const gameDate = new Date(now)
  if (now.getHours() < 4) {
    gameDate.setDate(gameDate.getDate() - 1)
  }
  return formatLocalDate(gameDate)
}

const getNextGameDateSwitchTime = () => {
  const now = new Date()
  const next = new Date(now)
  next.setHours(4, 0, 0, 0)
  if (now >= next) {
    next.setDate(next.getDate() + 1)
  }
  return next
}

const scheduleGameDateRefresh = () => {
  if (gameDateTimer) {
    clearTimeout(gameDateTimer)
  }
  const nextSwitch = getNextGameDateSwitchTime()
  const delay = Math.max(nextSwitch.getTime() - Date.now() + 1000, 1000)
  gameDateTimer = setTimeout(() => {
    const nextDate = getDefaultGameDate()
    if (!isDateManuallyEdited.value && form.date !== nextDate) {
      form.date = nextDate
      ElMessage.info(`已自动更新日期为 ${nextDate}`)
    }
    scheduleGameDateRefresh()
  }, delay)
}

const form = reactive({
  date: getDefaultGameDate(),
  player_id: '',
  gold_tubes: 0,
  purple_tubes: 0,
  sola_level: 8,
  claim_count: 1 as ClaimCount,
})

type ClaimCount = 1 | 2

type TacetCombo = {
  key: string
  label: string
  gold: number
  purple: number
  experience: number
}

const selectedComboKey = ref('')

const buildSingleCombos = (level: number): TacetCombo[] => {
  const combos = combosByLevel[level] || []
  return combos.map((combo) => ({
    key: `${combo.gold}-${combo.purple}`,
    label: `金${combo.gold}|紫${combo.purple}`,
    ...combo
  }))
}

const buildDoubleCombos = (level: number): TacetCombo[] => {
  const sourceCombos = combosByLevel[level] || []
  const comboMap = new Map<string, TacetCombo>()

  sourceCombos.forEach((leftCombo) => {
    sourceCombos.forEach((rightCombo) => {
      const gold = leftCombo.gold + rightCombo.gold
      const purple = leftCombo.purple + rightCombo.purple
      const key = `${gold}-${purple}`
      if (!comboMap.has(key)) {
        comboMap.set(key, {
          key,
          label: `金${gold}|紫${purple}`,
          gold,
          purple,
          experience: gold * 5000 + purple * 2000
        })
      }
    })
  })

  return Array.from(comboMap.values()).sort((a, b) => {
    if (b.gold !== a.gold) {
      return b.gold - a.gold
    }
    return b.purple - a.purple
  })
}

const availableCombos = computed<TacetCombo[]>(() => {
  return form.claim_count === 1
    ? buildSingleCombos(form.sola_level)
    : buildDoubleCombos(form.sola_level)
})

const currentCombo = computed(() =>
  availableCombos.value.find((combo) => combo.key === selectedComboKey.value) || null
)

const getDefaultComboKey = (level: number, claimCount: ClaimCount) => {
  const combos = claimCount === 1 ? buildSingleCombos(level) : buildDoubleCombos(level)
  if (level === 8) {
    if (claimCount === 1) {
      const preferred = combos.find((combo) => combo.gold === 3 && combo.purple === 4)
      if (preferred) {
        return preferred.key
      }
    } else {
      const preferred = combos.find((combo) => combo.gold === 7 && combo.purple === 8)
      if (preferred) {
        return preferred.key
      }
    }
  }
  return combos[0]?.key || ''
}

const applyComboToForm = () => {
  const combo = currentCombo.value || availableCombos.value[0]
  if (combo) {
    form.gold_tubes = combo.gold
    form.purple_tubes = combo.purple
  } else {
    form.gold_tubes = 0
    form.purple_tubes = 0
  }
}

const handleLevelChange = () => {
  selectedComboKey.value = getDefaultComboKey(form.sola_level, form.claim_count)
  applyComboToForm()
}

const handleClaimCountChange = (claimCount: ClaimCount) => {
  if (form.claim_count === claimCount) {
    return
  }
  form.claim_count = claimCount
  selectedComboKey.value = getDefaultComboKey(form.sola_level, form.claim_count)
  applyComboToForm()
}

const handleComboChange = () => {
  applyComboToForm()
}

const handleDateChange = () => {
  isDateManuallyEdited.value = true
}

const handleSubmit = async () => {
  if (!form.player_id) {
    ElMessage.warning('请输入玩家ID')
    return
  }

  loading.value = true
  try {
    const records = [{
      date: form.date,
      player_id: form.player_id,
      gold_tubes: form.gold_tubes,
      purple_tubes: form.purple_tubes,
      claim_count: form.claim_count,
      sola_level: form.sola_level
    }]

    await recordApi.createRecords(records)
    // 保存玩家ID到localStorage
    savePlayerId(form.player_id)
    ElMessage.success('录入成功')
    emit('success')
    handleReset()
  } catch (error) {
    ElMessage.error('录入失败: ' + (error as Error).message)
  } finally {
    loading.value = false
  }
}

const queryPlayerIds = (queryString: string, cb: (results: { value: string }[]) => void) => {
  const results = playerIds.value
    .filter((id) => id.toLowerCase().includes(queryString.toLowerCase()))
    .slice(0, 10)
    .map((id) => ({ value: id }))

  // 如果没有匹配，但用户有输入，返回用户的输入作为建议
  if (queryString && results.length === 0) {
    results.push({ value: queryString })
  }
  cb(results)
}

const loadPlayerIds = async () => {
  try {
    const response = await recordApi.getPlayerIds()
    playerIds.value = response.data
  } catch (error) {
    console.error('加载玩家ID列表失败:', error)
  }
}

const loadLastPlayerId = async () => {
  // 优先从localStorage获取
  const stored = getStoredPlayerId()
  if (stored) {
    form.player_id = stored
    return
  }

  // 备用：从服务器获取最近录入的玩家ID
  try {
    const response = await recordApi.getRecords({ limit: 1 })
    if (response.data.data.length > 0) {
      form.player_id = response.data.data[0].player_id
    }
  } catch (error) {
    console.error('获取最近玩家ID失败:', error)
  }
}

const handleReset = () => {
  form.date = getDefaultGameDate()
  isDateManuallyEdited.value = false
  form.sola_level = 8
  form.claim_count = 1
  selectedComboKey.value = getDefaultComboKey(form.sola_level, form.claim_count)
  applyComboToForm()
}

// 初始化
onMounted(async () => {
  handleLevelChange()
  scheduleGameDateRefresh()
  loadPlayerIds()
  await loadLastPlayerId()
})

onBeforeUnmount(() => {
  if (gameDateTimer) {
    clearTimeout(gameDateTimer)
    gameDateTimer = null
  }
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.option-button-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.combo-hint {
  margin-left: 12px;
  font-size: 13px;
}

.exp-hint {
  margin-left: 12px;
  color: #606266;
  font-size: 13px;
}

.material-gold {
  color: #b8860b;
  font-weight: 600;
}

.material-purple {
  color: #7d3c98;
  font-weight: 600;
}

:deep(.material-item-gold .el-form-item__label),
:deep(.material-item-gold .el-input-number__decrease),
:deep(.material-item-gold .el-input-number__increase),
:deep(.material-item-gold .el-input-number__input) {
  color: #b8860b;
  font-weight: 600;
}

:deep(.material-item-purple .el-form-item__label),
:deep(.material-item-purple .el-input-number__decrease),
:deep(.material-item-purple .el-input-number__increase),
:deep(.material-item-purple .el-input-number__input) {
  color: #7d3c98;
  font-weight: 600;
}
</style>
