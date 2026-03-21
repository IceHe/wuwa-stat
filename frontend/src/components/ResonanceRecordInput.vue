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
            @click="form.sola_level = level"
          >
            等级 {{ level }}
          </el-button>
        </div>
      </el-form-item>

      <el-form-item label="金" class="material-item-gold">
        <template v-if="goldOptions.length > 0">
          <div class="option-button-group">
            <el-button
              v-for="value in goldOptions"
              :key="`gold-${value}`"
              :type="form.gold === value ? 'primary' : 'default'"
              @click="form.gold = value"
            >
              {{ value }}
            </el-button>
          </div>
        </template>
        <el-input-number v-else v-model="form.gold" :min="0" />
      </el-form-item>

      <el-form-item label="紫" class="material-item-purple">
        <template v-if="purpleOptions.length > 0">
          <div class="option-button-group">
            <el-button
              v-for="value in purpleOptions"
              :key="`purple-${value}`"
              :type="form.purple === value ? 'primary' : 'default'"
              @click="form.purple = value"
            >
              {{ value }}
            </el-button>
          </div>
        </template>
        <el-input-number v-else v-model="form.purple" :min="0" />
      </el-form-item>

      <el-form-item label="蓝" class="material-item-blue">
        <template v-if="blueOptions.length > 0">
          <div class="option-button-group">
            <el-button
              v-for="value in blueOptions"
              :key="`blue-${value}`"
              :type="form.blue === value ? 'primary' : 'default'"
              @click="form.blue = value"
            >
              {{ value }}
            </el-button>
          </div>
        </template>
        <el-input-number v-else v-model="form.blue" :min="0" />
      </el-form-item>

      <el-form-item label="绿" class="material-item-green">
        <template v-if="greenOptions.length > 0">
          <div class="option-button-group">
            <el-button
              v-for="value in greenOptions"
              :key="`green-${value}`"
              :type="form.green === value ? 'primary' : 'default'"
              @click="form.green = value"
            >
              {{ value }}
            </el-button>
          </div>
        </template>
        <el-input-number v-else v-model="form.green" :min="0" />
      </el-form-item>

      <el-form-item v-if="presetHint" label="预设范围">
        <el-text type="info" size="small">{{ presetHint }}</el-text>
      </el-form-item>

      <el-form-item label="录入次数">
        <el-input-number
          v-model="form.count"
          :min="1"
          :max="10"
        />
        <el-text type="info" size="small" style="margin-left: 10px">
          相同数据录入多次（例如双倍领取）
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
import { ref, reactive, onMounted, onBeforeUnmount, computed, watch } from 'vue'
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
  gold: 0,
  purple: 0,
  blue: 0,
  green: 0,
  count: 1
})

type ResonanceDropPreset = {
  gold: number[]
  purple: number[]
  blue: number[]
  green: number[]
}

const dropPresetsByLevel = ref<Record<number, ResonanceDropPreset>>({})

const getSolaPreset = (level: number): ResonanceDropPreset => {
  return dropPresetsByLevel.value[level] || { gold: [], purple: [], blue: [], green: [] }
}

const goldOptions = computed(() => getSolaPreset(form.sola_level).gold)
const purpleOptions = computed(() => getSolaPreset(form.sola_level).purple)
const blueOptions = computed(() => getSolaPreset(form.sola_level).blue)
const greenOptions = computed(() => getSolaPreset(form.sola_level).green)

const formatOptionHint = (label: string, options: number[]) => {
  if (options.length === 0) return ''
  if (options.length === 1) return `${label}${options[0]}`
  const isContinuous = options.every((value, index) => index === 0 || value === options[index - 1] + 1)
  if (isContinuous) {
    return `${label}${options[0]}~${options[options.length - 1]}`
  }
  return `${label}${options.join('/')}`
}

const presetHint = computed(() => {
  const parts = [
    formatOptionHint('金', goldOptions.value),
    formatOptionHint('紫', purpleOptions.value),
    formatOptionHint('蓝', blueOptions.value),
    formatOptionHint('绿', greenOptions.value)
  ].filter(Boolean)

  return parts.length > 0 ? `索拉等级 ${form.sola_level}：${parts.join('，')}` : ''
})

const applySolaLevelPreset = () => {
  if (goldOptions.value.length > 0 && !goldOptions.value.includes(form.gold)) {
    form.gold = goldOptions.value[0]
  }
  if (purpleOptions.value.length > 0 && !purpleOptions.value.includes(form.purple)) {
    form.purple = purpleOptions.value[0]
  }
  if (blueOptions.value.length > 0 && !blueOptions.value.includes(form.blue)) {
    form.blue = blueOptions.value[0]
  }
  if (greenOptions.value.length > 0 && !greenOptions.value.includes(form.green)) {
    form.green = greenOptions.value[0]
  }
}

watch(
  () => form.sola_level,
  () => {
    applySolaLevelPreset()
  },
  { immediate: true }
)

const buildPresetValues = (values: number[]) => {
  return Array.from(new Set(values)).sort((a, b) => a - b)
}

const applyPresetOverrides = (level: number, preset: ResonanceDropPreset): ResonanceDropPreset => {
  if (level === 8) {
    return {
      ...preset,
      blue: [8]
    }
  }
  return preset
}

const loadDropPresetsFromData = async () => {
  try {
    const response = await resonanceApi.getDetailedStats()
    const levelStats = response.data.level_stats || []
    const presets: Record<number, ResonanceDropPreset> = {}

    levelStats.forEach((levelStat) => {
      const combinations = levelStat.combinations || []
      const rawPreset: ResonanceDropPreset = {
        gold: buildPresetValues(combinations.map((combo) => combo.gold)),
        purple: buildPresetValues(combinations.map((combo) => combo.purple)),
        blue: buildPresetValues(combinations.map((combo) => combo.blue)),
        green: buildPresetValues(combinations.map((combo) => combo.green))
      }

      presets[levelStat.sola_level] = applyPresetOverrides(levelStat.sola_level, rawPreset)
    })

    dropPresetsByLevel.value = presets
    applySolaLevelPreset()
  } catch (error) {
    console.error('加载凝素领域预设失败:', error)
  }
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
    const records = Array(form.count).fill(null).map(() => ({
      date: form.date,
      player_id: form.player_id,
      sola_level: form.sola_level,
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
  form.gold = 0
  form.purple = 0
  form.blue = 0
  form.green = 0
  applySolaLevelPreset()
  form.count = 1
}

onMounted(async () => {
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
