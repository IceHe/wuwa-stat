<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span>批量录入突破材料掉落记录</span>
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
        <el-select v-model="form.sola_level" placeholder="选择索拉等级">
          <el-option v-for="level in solaLevels" :key="level" :label="`等级 ${level}`" :value="level" />
        </el-select>
      </el-form-item>

      <el-form-item label="掉落数量">
        <el-select v-model="form.drop_count" placeholder="选择掉落数量" style="width: 100%">
          <el-option
            v-for="drop in dropCountOptions"
            :key="drop"
            :label="`掉落 ${drop}`"
            :value="drop"
          />
        </el-select>
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
import { ref, reactive, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import { ascensionApi } from '../api'

const emit = defineEmits(['success'])

const STORAGE_KEY = 'wuwa_last_ascension_player_id'
const playerIds = ref<string[]>([])

const getStoredPlayerId = (): string | null => {
  return localStorage.getItem(STORAGE_KEY)
}

const savePlayerId = (playerId: string) => {
  localStorage.setItem(STORAGE_KEY, playerId)
}

const solaLevels = [8, 7, 6, 5, 4, 3, 2, 1]

const dropCountOptionsByLevel: Record<number, number[]> = {
  8: [4, 5],
  7: [4, 5],
  6: [2, 3]
}

const getDropCountOptions = (solaLevel: number) => dropCountOptionsByLevel[solaLevel] ?? []

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
  drop_count: 4,
  count: 1
})

const dropCountOptions = computed(() => getDropCountOptions(form.sola_level))

watch(
  () => form.sola_level,
  (level) => {
    const options = getDropCountOptions(level)
    if (options.length > 0 && !options.includes(form.drop_count)) {
      form.drop_count = options[0]
    }
  },
  { immediate: true }
)

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
      drop_count: form.drop_count
    }))

    await ascensionApi.createRecords(records)
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
    const response = await ascensionApi.getPlayerIds()
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
    const response = await ascensionApi.getRecords({ limit: 1 })
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
  form.drop_count = getDropCountOptions(form.sola_level)[0] ?? 0
  form.count = 1
}

onMounted(async () => {
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
</style>
