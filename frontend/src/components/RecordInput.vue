<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span>批量录入产出记录</span>
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
        <el-select v-model="form.sola_level" placeholder="选择索拉等级" @change="handleLevelChange">
          <el-option v-for="level in solaLevels" :key="level" :label="`等级 ${level}`" :value="level" />
        </el-select>
      </el-form-item>

      <el-form-item label="掉落组合">
        <el-select
          v-model="selectedComboKey"
          placeholder="选择掉落组合"
          @change="handleComboChange"
        >
          <el-option
            v-for="combo in availableCombos"
            :key="combo.key"
            :label="combo.label"
            :value="combo.key"
          />
        </el-select>
        <div class="exp-hint" v-if="currentCombo">
          声骸经验：{{ currentCombo.experience.toLocaleString() }}
        </div>
      </el-form-item>

      <el-form-item label="金色密音筒">
        <el-input-number v-model="form.gold_tubes" :min="0" disabled />
      </el-form-item>

      <el-form-item label="紫色密音筒">
        <el-input-number v-model="form.purple_tubes" :min="0" disabled />
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
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import { recordApi } from '../api'

const emit = defineEmits(['success'])

const playerIds = ref<string[]>([])

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
  count: 1
})

const selectedComboKey = ref('')

const availableCombos = computed(() => {
  const combos = combosByLevel[form.sola_level] || []
  return combos.map((c) => ({
    key: `${c.gold}-${c.purple}`,
    label: `金 ${c.gold} 紫 ${c.purple}`,
    ...c
  }))
})

const currentCombo = computed(() =>
  availableCombos.value.find((c) => c.key === selectedComboKey.value)
)

const getDefaultComboKey = (level: number) => {
  const combos = combosByLevel[level] || []
  if (level === 8) {
    const preferred = combos.find((c) => c.gold === 3 && c.purple === 4)
    if (preferred) {
      return `${preferred.gold}-${preferred.purple}`
    }
  }
  const first = combos[0]
  return first ? `${first.gold}-${first.purple}` : ''
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
  selectedComboKey.value = getDefaultComboKey(form.sola_level)
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
    const records = Array(form.count).fill(null).map(() => ({
      date: form.date,
      player_id: form.player_id,
      gold_tubes: form.gold_tubes,
      purple_tubes: form.purple_tubes,
      sola_level: form.sola_level
    }))

    await recordApi.createRecords(records)
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
  // 从服务器获取最近录入的玩家ID
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
  form.count = 1
  selectedComboKey.value = getDefaultComboKey(form.sola_level)
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

.exp-hint {
  margin-left: 12px;
  color: #606266;
  font-size: 13px;
}
</style>
