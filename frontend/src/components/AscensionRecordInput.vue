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
            @click="form.sola_level = level"
          >
            等级 {{ level }}
          </el-button>
        </div>
      </el-form-item>

      <el-form-item label="掉落数量">
        <div class="option-button-group">
          <el-button
            v-for="drop in dropCountOptions"
            :key="drop"
            :type="form.drop_count === drop ? 'primary' : 'default'"
            @click="form.drop_count = drop"
          >
            掉落 {{ drop }}
          </el-button>
        </div>
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
import { confirmRecordWithoutEnergyDeduction, isEnergyInsufficientError } from '../utils/energyDeduction'
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
  refreshAccount: () => Promise<void>
}

const playerIds = ref<string[]>([])
const playerIdFieldRef = ref<PlayerIdFieldExpose | null>(null)

const solaLevels = [8, 7, 6, 5, 4, 3, 2, 1]

const dropCountOptionsByLevel: Record<number, number[]> = {
  8: [4, 5],
  7: [4, 5],
  6: [2, 3],
  5: [2, 3],
  4: [1, 2],
  3: [1, 2],
  2: [1, 2],
  1: [1, 2]
}

const getDropCountOptions = (solaLevel: number) => dropCountOptionsByLevel[solaLevel] ?? []
const getDefaultDropCount = (solaLevel: number) => getDropCountOptions(solaLevel)[0] ?? 0

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
  sola_level: 8,
  drop_count: 4,
  count: 1
})

const dropCountOptions = computed(() => getDropCountOptions(form.sola_level))

watch(
  () => form.sola_level,
  (level) => {
    form.drop_count = getDefaultDropCount(level)
  },
  { immediate: true }
)

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

const submitRecords = async (skipEnergyDeduction = false) => {
  const submittedSolaLevel = form.sola_level
  const submittedCount = form.count
  const records = Array(submittedCount).fill(null).map(() => ({
    date: form.date,
    player_id: form.player_id,
    sola_level: submittedSolaLevel,
    drop_count: form.drop_count
  }))

  await ascensionApi.createRecords(records, { skipEnergyDeduction })
  void playerIdFieldRef.value?.refreshAccount()
  emit('update:playerId', form.player_id)
  ElMessage.success(`成功录入 ${submittedCount} 条记录`)
  emit('success')
  resetForm(submittedSolaLevel)
}

const handleSubmit = async () => {
  if (!form.player_id) {
    ElMessage.warning('请输入玩家ID')
    return
  }

  loading.value = true
  try {
    await submitRecords()
  } catch (error) {
    if (!isEnergyInsufficientError(error)) {
      ElMessage.error('录入失败: ' + (error as Error).message)
      return
    }

    loading.value = false
    const confirmed = await confirmRecordWithoutEnergyDeduction()
    if (!confirmed) {
      return
    }

    loading.value = true
    try {
      await submitRecords(true)
    } catch (retryError) {
      ElMessage.error('录入失败: ' + (retryError as Error).message)
    }
  } finally {
    loading.value = false
  }
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
  if (props.playerId) {
    form.player_id = props.playerId
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

const resetForm = (solaLevel: number) => {
  form.date = getDefaultGameDate()
  isDateManuallyEdited.value = false
  form.sola_level = solaLevel
  form.drop_count = getDefaultDropCount(form.sola_level)
  form.count = 1
}

const handleReset = () => {
  resetForm(8)
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

.option-button-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
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
