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

      <el-form-item>
        <template #label>
          <button type="button" class="player-id-label-button" @click="openPlayerIdDialog">
            玩家ID
          </button>
        </template>
        <PlayerIdField ref="playerIdFieldRef" v-model="form.player_id" :player-ids="playerIds" />
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

      <el-form-item label="金色密音筒" class="material-item-gold">
        <div class="option-button-group">
          <el-button
            v-for="value in goldOptions"
            :key="`gold-${value}`"
            :type="form.gold_tubes === value ? 'primary' : 'default'"
            @click="handleGoldChange(value)"
          >
            {{ value }}
          </el-button>
        </div>
      </el-form-item>

      <el-form-item label="紫色密音筒" class="material-item-purple">
        <div class="option-button-group">
          <el-button
            v-for="value in purpleOptions"
            :key="`purple-${value}`"
            :type="form.purple_tubes === value ? 'primary' : 'default'"
            @click="handlePurpleChange(value)"
          >
            {{ value }}
          </el-button>
        </div>
        <div class="combo-hint" v-if="currentCombo">
          <span class="material-gold">金{{ currentCombo.gold }}</span>
          <span> </span>
          <span class="material-purple">紫{{ currentCombo.purple }}</span>
        </div>
        <div class="exp-hint">
          {{ form.claim_count === 1 ? '单次领取合法组合范围' : '两次领取合并后的合法组合范围' }}
        </div>
        <div class="exp-hint" v-if="currentCombo">
          声骸经验：{{ currentCombo.experience.toLocaleString() }}
        </div>
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
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { tacetApi } from '../api'
import PlayerIdField from './PlayerIdField.vue'

const props = withDefaults(defineProps<{ playerId?: string }>(), {
  playerId: ''
})

const emit = defineEmits<{
  (e: 'success'): void
  (e: 'update:playerId', value: string): void
}>()

type PlayerIdFieldExpose = {
  openDialog: () => void
}

const playerIds = ref<string[]>([])
const playerIdFieldRef = ref<PlayerIdFieldExpose | null>(null)

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
  player_id: props.playerId,
  gold_tubes: 0,
  purple_tubes: 0,
  sola_level: 8,
  claim_count: 1 as ClaimCount,
})

type ClaimCount = 1 | 2

type TacetCombo = {
  key: string
  gold: number
  purple: number
  experience: number
}

const buildSingleCombos = (level: number): TacetCombo[] => {
  const combos = combosByLevel[level] || []
  return combos.map((combo) => ({
    key: `${combo.gold}-${combo.purple}`,
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
          gold,
          purple,
          experience: gold * 5000 + purple * 2000
        })
      }
    })
  })

  return Array.from(comboMap.values()).sort((a, b) => {
    if (a.purple !== b.purple) {
      return a.purple - b.purple
    }
    return a.gold - b.gold
  })
}

const availableCombos = computed<TacetCombo[]>(() => {
  return form.claim_count === 1
    ? buildSingleCombos(form.sola_level)
    : buildDoubleCombos(form.sola_level)
})

const currentCombo = computed(() =>
  availableCombos.value.find(
    (combo) => combo.gold === form.gold_tubes && combo.purple === form.purple_tubes
  ) || null
)

const getDefaultCombo = (level: number, claimCount: ClaimCount) => {
  const combos = claimCount === 1 ? buildSingleCombos(level) : buildDoubleCombos(level)
  if (level === 8) {
    if (claimCount === 1) {
      const preferred = combos.find((combo) => combo.gold === 3 && combo.purple === 4)
      if (preferred) {
        return preferred
      }
    } else {
      const preferred = combos.find((combo) => combo.gold === 6 && combo.purple === 8)
      if (preferred) {
        return preferred
      }
    }
  }
  return combos[0] || null
}

const getUniqueSortedValues = (values: number[]) => Array.from(new Set(values)).sort((a, b) => a - b)

const purpleOptions = computed(() => {
  const combos = availableCombos.value
  const combosForGold = combos.filter((combo) => combo.gold === form.gold_tubes)
  const source = combosForGold.length > 0 ? combosForGold : combos
  return getUniqueSortedValues(source.map((combo) => combo.purple))
})

const goldOptions = computed(() => {
  const combos = availableCombos.value
  const combosForPurple = combos.filter((combo) => combo.purple === form.purple_tubes)
  const source = combosForPurple.length > 0 ? combosForPurple : combos
  return getUniqueSortedValues(source.map((combo) => combo.gold))
})

const applyComboToForm = (combo?: TacetCombo | null) => {
  const nextCombo = combo || currentCombo.value || availableCombos.value[0]
  if (nextCombo) {
    form.gold_tubes = nextCombo.gold
    form.purple_tubes = nextCombo.purple
  } else {
    form.gold_tubes = 0
    form.purple_tubes = 0
  }
}

const ensureValidSelection = () => {
  const combo = currentCombo.value
  if (combo) {
    form.gold_tubes = combo.gold
    form.purple_tubes = combo.purple
    return
  }

  applyComboToForm(getDefaultCombo(form.sola_level, form.claim_count))
}

const handleLevelChange = () => {
  applyComboToForm(getDefaultCombo(form.sola_level, form.claim_count))
}

const handleClaimCountChange = (claimCount: ClaimCount) => {
  if (form.claim_count === claimCount) {
    return
  }
  form.claim_count = claimCount
  applyComboToForm(getDefaultCombo(form.sola_level, form.claim_count))
}

const handleGoldChange = (value: number) => {
  if (form.gold_tubes === value) {
    return
  }

  form.gold_tubes = value
  const matchingCombo = availableCombos.value.find(
    (combo) => combo.gold === value && combo.purple === form.purple_tubes
  )
  if (matchingCombo) {
    applyComboToForm(matchingCombo)
    return
  }

  const fallbackCombo = availableCombos.value.find((combo) => combo.gold === value)
  applyComboToForm(fallbackCombo)
}

const handlePurpleChange = (value: number) => {
  if (form.purple_tubes === value) {
    return
  }

  form.purple_tubes = value
  const matchingCombo = availableCombos.value.find(
    (combo) => combo.purple === value && combo.gold === form.gold_tubes
  )
  if (matchingCombo) {
    applyComboToForm(matchingCombo)
    return
  }

  const fallbackCombo = availableCombos.value.find((combo) => combo.purple === value)
  applyComboToForm(fallbackCombo)
}

const handleDateChange = () => {
  isDateManuallyEdited.value = true
}

const openPlayerIdDialog = () => {
  playerIdFieldRef.value?.openDialog()
}

watch(
  () => props.playerId,
  (playerId) => {
    if (form.player_id !== playerId) {
      form.player_id = playerId
    }
  }
)

watch(
  () => form.player_id,
  (playerId) => {
    if (playerId !== props.playerId) {
      emit('update:playerId', playerId)
    }
  }
)

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

    await tacetApi.createRecords(records)
    emit('update:playerId', form.player_id)
    ElMessage.success('录入成功')
    emit('success')
    handleReset()
  } catch (error) {
    ElMessage.error('录入失败: ' + (error as Error).message)
  } finally {
    loading.value = false
  }
}

const loadPlayerIds = async () => {
  try {
    const response = await tacetApi.getPlayerIds()
    playerIds.value = response.data
  } catch (error) {
    console.error('加载玩家ID列表失败:', error)
  }
}

const loadLastPlayerId = async () => {
  if (props.playerId) {
    form.player_id = props.playerId
    return
  }

  // 备用：从服务器获取最近录入的玩家ID
  try {
    const response = await tacetApi.getRecords({ limit: 1 })
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
  applyComboToForm(getDefaultCombo(form.sola_level, form.claim_count))
}

// 初始化
onMounted(async () => {
  ensureValidSelection()
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

.player-id-label-button {
  appearance: none;
  background: transparent;
  border: 0;
  color: inherit;
  cursor: pointer;
  font: inherit;
  padding: 0;
}

.player-id-label-button:hover,
.player-id-label-button:focus-visible {
  color: #409eff;
}

.player-id-label-button:focus-visible {
  border-radius: 2px;
  outline: 2px solid #409eff;
  outline-offset: 2px;
}
</style>
