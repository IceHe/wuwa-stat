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
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { recordApi } from '../api'

const route = useRoute()
const router = useRouter()

const emit = defineEmits(['success'])

const playerIds = ref<string[]>([])
const isInitialized = ref(false)

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
const form = reactive({
  date: new Date().toISOString().split('T')[0],
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
  // 切换等级时默认选中第一组组合
  selectedComboKey.value = availableCombos.value[0]?.key || ''
  applyComboToForm()
}

const handleComboChange = () => {
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
  // 优先使用URL参数
  const urlPlayerId = route.query.player_id as string
  if (urlPlayerId) {
    form.player_id = urlPlayerId
    return
  }

  // 从服务器获取最近录入的玩家ID
  try {
    const response = await recordApi.getRecords({ limit: 1 })
    if (response.data.length > 0) {
      form.player_id = response.data[0].player_id
    }
  } catch (error) {
    console.error('获取最近玩家ID失败:', error)
  }
}

// 监听玩家ID变化，同步到URL
watch(() => form.player_id, (newVal) => {
  if (!isInitialized.value) return
  const query = { ...route.query }
  if (newVal) {
    query.player_id = newVal
  } else {
    delete query.player_id
  }
  router.replace({ query })
})

const handleReset = () => {
  form.date = new Date().toISOString().split('T')[0]
  form.sola_level = 8
  form.count = 1
  selectedComboKey.value = availableCombos.value[0]?.key || ''
  applyComboToForm()
}

// 初始化
onMounted(async () => {
  handleLevelChange()
  loadPlayerIds()
  await loadLastPlayerId()
  isInitialized.value = true
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