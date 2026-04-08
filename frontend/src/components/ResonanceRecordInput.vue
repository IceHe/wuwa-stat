<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span>批量录入凝素领域产出记录</span>
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
        <div class="option-button-group" v-if="availableCombos.length > 0">
          <el-button
            v-for="combo in availableCombos"
            :key="combo.key"
            :type="selectedComboKey === combo.key ? 'primary' : 'default'"
            @click="selectedComboKey !== combo.key && ((selectedComboKey = combo.key), handleComboChange())"
          >
            <span class="material-gold">金{{ combo.gold }}</span>
            <span> </span>
            <span class="material-purple">紫{{ combo.purple }}</span>
            <span> </span>
            <span class="material-blue">蓝{{ combo.blue }}</span>
            <span> </span>
            <span class="material-green">绿{{ combo.green }}</span>
          </el-button>
        </div>
        <div v-else class="exp-hint">
          当前索拉等级暂无可选组合，请手动补录数据后再使用组合快捷录入。
        </div>
        <div class="combo-hint" v-if="currentCombo">
          <span class="material-gold">金{{ currentCombo.gold }}</span>
          <span> </span>
          <span class="material-purple">紫{{ currentCombo.purple }}</span>
          <span> </span>
          <span class="material-blue">蓝{{ currentCombo.blue }}</span>
          <span> </span>
          <span class="material-green">绿{{ currentCombo.green }}</span>
        </div>
        <div class="exp-hint" v-if="currentCombo">
          {{ form.claim_count === 1 ? '单次领取组合' : '两次领取合并后的组合' }}
        </div>
      </el-form-item>

      <el-form-item label="金" class="material-item-gold">
        <el-input-number v-model="form.gold" :min="0" :disabled="availableCombos.length > 0" />
      </el-form-item>

      <el-form-item label="紫" class="material-item-purple">
        <el-input-number v-model="form.purple" :min="0" :disabled="availableCombos.length > 0" />
      </el-form-item>

      <el-form-item label="蓝" class="material-item-blue">
        <el-input-number v-model="form.blue" :min="0" :disabled="availableCombos.length > 0" />
      </el-form-item>

      <el-form-item label="绿" class="material-item-green">
        <el-input-number v-model="form.green" :min="0" :disabled="availableCombos.length > 0" />
      </el-form-item>

      <el-form-item label="录入条数">
        <el-input-number
          v-model="form.count"
          :min="1"
          :max="10"
        />
        <el-text type="info" size="small" style="margin-left: 10px">
          仅用于连续录入多条完全相同的记录
        </el-text>
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
import { ref, reactive, onMounted, onBeforeUnmount, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { resonanceApi } from '../api'

const emit = defineEmits(['success'])

const STORAGE_KEY = 'wuwa_last_resonance_player_id'
const playerIds = ref<string[]>([])

const getStoredPlayerId = (): string | null => {
  return localStorage.getItem(STORAGE_KEY)
}

const savePlayerId = (playerId: string) => {
  localStorage.setItem(STORAGE_KEY, playerId)
}

const solaLevels = [8, 7, 6, 5, 4, 3, 2, 1]

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
  sola_level: 8,
  claim_count: 1 as ClaimCount,
  gold: 0,
  purple: 0,
  blue: 0,
  green: 0,
  count: 1
})

type ClaimCount = 1 | 2

type ResonanceDropCombination = {
  claim_count: number
  gold: number
  purple: number
  blue: number
  green: number
  count?: number
}

type ResonanceCombo = {
  key: string
  gold: number
  purple: number
  blue: number
  green: number
  count?: number
}

const selectedComboKey = ref('')

const combosByLevel = ref<Record<number, Record<ClaimCount, ResonanceCombo[]>>>({})

const buildComboKey = (combo: Pick<ResonanceCombo, 'gold' | 'purple' | 'blue' | 'green'>) => {
  return [combo.gold, combo.purple, combo.blue, combo.green].join('-')
}

const sortCombos = (combos: ResonanceCombo[]) => {
  return [...combos].sort((left, right) => {
    if ((right.count || 0) !== (left.count || 0)) {
      return (right.count || 0) - (left.count || 0)
    }
    if (right.gold !== left.gold) return right.gold - left.gold
    if (right.purple !== left.purple) return right.purple - left.purple
    if (right.blue !== left.blue) return right.blue - left.blue
    return right.green - left.green
  })
}

const availableCombos = computed(() => combosByLevel.value[form.sola_level]?.[form.claim_count] || [])

const currentCombo = computed(() => {
  return availableCombos.value.find((combo) => combo.key === selectedComboKey.value) || null
})

const buildDoubleCombinations = (singleClaimCombos: ResonanceDropCombination[]) => {
  const combinationMap = new Map<string, ResonanceCombo>()

  singleClaimCombos.forEach((leftCombo) => {
    singleClaimCombos.forEach((rightCombo) => {
      const combo: ResonanceCombo = {
        key: '',
        gold: leftCombo.gold + rightCombo.gold,
        purple: leftCombo.purple + rightCombo.purple,
        blue: leftCombo.blue + rightCombo.blue,
        green: leftCombo.green + rightCombo.green
      }
      const key = buildComboKey(combo)
      combo.key = key

      if (!combinationMap.has(key)) {
        combinationMap.set(key, combo)
      }
    })
  })

  return sortCombos(Array.from(combinationMap.values()))
}

const buildCombos = (combinations: ResonanceDropCombination[]) => {
  const comboMap = new Map<string, ResonanceCombo>()

  combinations.forEach((combo) => {
    const key = buildComboKey(combo)
    const existing = comboMap.get(key)
    if (existing) {
      existing.count = (existing.count || 0) + (combo.count || 0)
      return
    }
    comboMap.set(key, {
      key,
      gold: combo.gold,
      purple: combo.purple,
      blue: combo.blue,
      green: combo.green,
      count: combo.count || 0
    })
  })

  return sortCombos(Array.from(comboMap.values()))
}

const applyLevelComboOverrides = (level: number, claimCount: ClaimCount, combos: ResonanceCombo[]) => {
  if (level === 8) {
    const requiredBlue = claimCount === 1 ? 8 : 16
    const filteredCombos = combos.filter((combo) => combo.blue === requiredBlue)
    if (filteredCombos.length > 0) {
      return filteredCombos
    }
  }

  return combos
}

const getDefaultComboKey = (level: number, claimCount: ClaimCount) => {
  const combos = combosByLevel.value[level]?.[claimCount] || []
  return combos[0]?.key || ''
}

const applyComboToForm = () => {
  const combo = currentCombo.value || availableCombos.value[0]
  if (!combo) {
    form.gold = 0
    form.purple = 0
    form.blue = 0
    form.green = 0
    return
  }

  if (selectedComboKey.value !== combo.key) {
    selectedComboKey.value = combo.key
  }

  form.gold = combo.gold
  form.purple = combo.purple
  form.blue = combo.blue
  form.green = combo.green
}

const handleLevelChange = () => {
  selectedComboKey.value = getDefaultComboKey(form.sola_level, form.claim_count)
  applyComboToForm()
}

const handleComboChange = () => {
  applyComboToForm()
}

const loadDropPresetsFromData = async () => {
  try {
    const response = await resonanceApi.getDetailedStats()
    const levelStats = response.data.level_stats || []
    const nextCombosByLevel: Record<number, Record<ClaimCount, ResonanceCombo[]>> = {}

    levelStats.forEach((levelStat) => {
      const combinations = levelStat.combinations || []
      const singleClaimCombos = combinations.filter((combo) => combo.claim_count === 1)
      const doubleClaimCombos = combinations.filter((combo) => combo.claim_count === 2)
      const singleCombos = buildCombos(singleClaimCombos)
      const explicitDoubleCombos = buildCombos(doubleClaimCombos)
      const inferredDoubleCombos = buildDoubleCombinations(singleClaimCombos)
      const mergedDoubleCombos = buildCombos([
        ...doubleClaimCombos,
        ...inferredDoubleCombos.map((combo) => ({
          claim_count: 2,
          gold: combo.gold,
          purple: combo.purple,
          blue: combo.blue,
          green: combo.green,
          count: combo.count || 0
        }))
      ])

      nextCombosByLevel[levelStat.sola_level] = {
        1: applyLevelComboOverrides(levelStat.sola_level, 1, singleCombos),
        2: applyLevelComboOverrides(
          levelStat.sola_level,
          2,
          mergedDoubleCombos.length > 0 ? mergedDoubleCombos : explicitDoubleCombos
        )
      }
    })

    combosByLevel.value = nextCombosByLevel
    handleLevelChange()
  } catch (error) {
    console.error('加载凝素领域预设失败:', error)
  }
}

const handleDateChange = () => {
  isDateManuallyEdited.value = true
}

const handleClaimCountChange = (claimCount: ClaimCount) => {
  if (form.claim_count === claimCount) {
    return
  }
  form.claim_count = claimCount
  selectedComboKey.value = getDefaultComboKey(form.sola_level, form.claim_count)
  applyComboToForm()
}

const handleSubmit = async () => {
  if (!form.player_id) {
    ElMessage.warning('请输入玩家ID')
    return
  }

  loading.value = true
  try {
    const records = Array(form.count).fill(null).map(() => ({
      date: form.date,
      player_id: form.player_id,
      sola_level: form.sola_level,
      claim_count: form.claim_count,
      gold: form.gold,
      purple: form.purple,
      blue: form.blue,
      green: form.green
    }))

    await resonanceApi.createRecords(records)
    savePlayerId(form.player_id)
    ElMessage.success(`成功录入 ${form.count} 条记录`)
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

  if (queryString && results.length === 0) {
    results.push({ value: queryString })
  }
  cb(results)
}

const loadPlayerIds = async () => {
  try {
    const response = await resonanceApi.getPlayerIds()
    playerIds.value = response.data
  } catch (error) {
    console.error('加载玩家ID列表失败:', error)
  }
}

const loadLastPlayerId = async () => {
  const stored = getStoredPlayerId()
  if (stored) {
    form.player_id = stored
    return
  }

  try {
    const response = await resonanceApi.getRecords({ limit: 1 })
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
  form.count = 1
}

onMounted(async () => {
  handleLevelChange()
  scheduleGameDateRefresh()
  loadPlayerIds()
  await Promise.all([
    loadLastPlayerId(),
    loadDropPresetsFromData()
  ])
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

.material-blue {
  color: #1f78ff;
  font-weight: 600;
}

.material-green {
  color: #2e8b57;
  font-weight: 600;
}

:deep(.material-item-gold .el-form-item__label),
:deep(.material-item-gold .el-input-number__decrease),
:deep(.material-item-gold .el-input-number__increase),
:deep(.material-item-gold .el-input-number__input),
:deep(.material-item-gold .el-select .el-input__inner) {
  color: #b8860b;
  font-weight: 600;
}

:deep(.material-item-purple .el-form-item__label),
:deep(.material-item-purple .el-input-number__decrease),
:deep(.material-item-purple .el-input-number__increase),
:deep(.material-item-purple .el-input-number__input),
:deep(.material-item-purple .el-select .el-input__inner) {
  color: #7d3c98;
  font-weight: 600;
}

:deep(.material-item-blue .el-form-item__label),
:deep(.material-item-blue .el-input-number__decrease),
:deep(.material-item-blue .el-input-number__increase),
:deep(.material-item-blue .el-input-number__input),
:deep(.material-item-blue .el-select .el-input__inner) {
  color: #1f78ff;
  font-weight: 600;
}

:deep(.material-item-green .el-form-item__label),
:deep(.material-item-green .el-input-number__decrease),
:deep(.material-item-green .el-input-number__increase),
:deep(.material-item-green .el-input-number__input),
:deep(.material-item-green .el-select .el-input__inner) {
  color: #2e8b57;
  font-weight: 600;
}
</style>
