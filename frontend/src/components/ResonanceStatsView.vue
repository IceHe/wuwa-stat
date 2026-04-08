<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span>凝素领域产出统计</span>
        <el-button type="primary" size="small" @click="loadStats">
          刷新
        </el-button>
      </div>
    </template>

    <el-form :inline="true" :model="filters">
      <el-form-item label="玩家ID">
        <el-input v-model="filters.player_id" placeholder="筛选玩家" clearable />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="loadStats">查询</el-button>
      </el-form-item>
    </el-form>

    <div v-loading="loading">
      <div v-if="!level8Summary" class="no-data">
        暂无索拉等级8数据
      </div>
      <div v-else>
        <div class="summary-title">凝素领域产出　索拉等级8</div>
        <el-table :data="level8Rows" border style="margin-bottom: 12px" :row-class-name="getMaterialRowClass">
          <el-table-column prop="label" label="" width="120" align="center" />
          <el-table-column prop="total" label="总数" width="140" align="center" />
          <el-table-column prop="avg" label="平均" width="140" align="center" />
          <el-table-column prop="combined" label="向上合成后合计" min-width="180" align="center" />
        </el-table>

        <el-descriptions :column="2" border style="margin-bottom: 16px">
          <el-descriptions-item label="统计记录数">{{ level8Summary.totalCount }}</el-descriptions-item>
          <el-descriptions-item label="统计领取次数">{{ level8Summary.totalClaimCount }}</el-descriptions-item>
          <el-descriptions-item label="双倍一周总领取次数">{{ weeklyDoubleRuns }}</el-descriptions-item>
        </el-descriptions>

        <div class="summary-title">角色0+1需求</div>
        <el-table :data="roleDemandRows" border style="margin-bottom: 12px" :row-class-name="getMaterialRowClass">
          <el-table-column prop="label" label="品质" width="120" align="center" />
          <el-table-column prop="character" label="90级角色技能升满10级" min-width="180" align="center" />
          <el-table-column prop="weapon" label="五星专武升满90级" min-width="170" align="center" />
          <el-table-column prop="total" label="90级角色0+1升满技能+专武" min-width="210" align="center" />
        </el-table>

        <el-table :data="doubleWeekRows" border style="margin-bottom: 12px" :row-class-name="getMaterialRowClass">
          <el-table-column prop="label" label="品质" width="120" align="center" />
          <el-table-column prop="weekly" label="索8产出" width="140" align="center" />
          <el-table-column prop="surplus" label="升满0+1盈余" width="160" align="center" />
          <el-table-column prop="remark" label="备注" min-width="280" align="center" />
        </el-table>

        <el-descriptions :column="1" border>
          <el-descriptions-item label="一周双倍后，升满0+1还需领几次凝素领域奖励？">
            {{ requiredRunsAfterDouble }}
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted, computed } from 'vue'
import { resonanceApi, type ResonanceDetailedStats } from '../api'
import { ElMessage } from 'element-plus'

const props = defineProps<{
  refresh: number
}>()

const loading = ref(false)
const detailedStats = ref<ResonanceDetailedStats>({
  level_stats: []
})

const filters = reactive({
  player_id: ''
})

const level8Summary = computed(() => {
  const level8 = detailedStats.value.level_stats.find((item) => item.sola_level === 8)
  if (!level8) return null

  const totalCount = level8.total_count
  const totalClaimCount = level8.total_claim_count
  const totalGold = level8.total_gold
  const totalPurple = level8.total_purple
  const totalBlue = level8.total_blue
  const totalGreen = level8.total_green

  const avgGold = totalClaimCount > 0 ? totalGold / totalClaimCount : 0
  const avgPurple = totalClaimCount > 0 ? totalPurple / totalClaimCount : 0
  const avgBlue = totalClaimCount > 0 ? totalBlue / totalClaimCount : 0
  const avgGreen = totalClaimCount > 0 ? totalGreen / totalClaimCount : 0

  const combinedGold = avgGold + avgPurple / 3 + avgBlue / 9 + avgGreen / 27
  const combinedPurple = avgPurple + avgBlue / 3 + avgGreen / 9
  const combinedBlue = avgBlue + avgGreen / 3
  const combinedGreen = avgGreen

  return {
    totalCount,
    totalClaimCount,
    totalGold,
    totalPurple,
    totalBlue,
    totalGreen,
    avgGold,
    avgPurple,
    avgBlue,
    avgGreen,
    combinedGold,
    combinedPurple,
    combinedBlue,
    combinedGreen
  }
})

const formatNumber = (value: number) => {
  return value.toFixed(2)
}

const level8Rows = computed(() => {
  if (!level8Summary.value) return []

  return [
    {
      label: '金',
      total: level8Summary.value.totalGold,
      avg: formatNumber(level8Summary.value.avgGold),
      combined: formatNumber(level8Summary.value.combinedGold)
    },
    {
      label: '紫',
      total: level8Summary.value.totalPurple,
      avg: formatNumber(level8Summary.value.avgPurple),
      combined: formatNumber(level8Summary.value.combinedPurple)
    },
    {
      label: '蓝',
      total: level8Summary.value.totalBlue,
      avg: formatNumber(level8Summary.value.avgBlue),
      combined: formatNumber(level8Summary.value.combinedBlue)
    },
    {
      label: '绿',
      total: level8Summary.value.totalGreen,
      avg: formatNumber(level8Summary.value.avgGreen),
      combined: formatNumber(level8Summary.value.combinedGreen)
    }
  ]
})

const weeklyDoubleRuns = 42

const roleDemand = {
  gold: { character: 67, weapon: 20, total: 87 },
  purple: { character: 55, weapon: 6, total: 61 },
  blue: { character: 28, weapon: 8, total: 36 },
  green: { character: 25, weapon: 6, total: 31 }
}

const roleDemandRows = [
  { label: '金', ...roleDemand.gold },
  { label: '紫', ...roleDemand.purple },
  { label: '蓝', ...roleDemand.blue },
  { label: '绿', ...roleDemand.green }
]

const doubleWeekSummary = computed(() => {
  if (!level8Summary.value) return null

  const weeklyGold = level8Summary.value.avgGold * weeklyDoubleRuns
  const weeklyPurple = level8Summary.value.avgPurple * weeklyDoubleRuns
  const weeklyBlue = level8Summary.value.avgBlue * weeklyDoubleRuns
  const weeklyGreen = level8Summary.value.avgGreen * weeklyDoubleRuns

  const greenSurplus = weeklyGreen - roleDemand.green.total
  const blueSurplus = weeklyBlue + greenSurplus / 3 - roleDemand.blue.total
  const purpleSurplus = weeklyPurple + blueSurplus / 3 - roleDemand.purple.total
  const goldSurplus = weeklyGold + purpleSurplus / 3 - roleDemand.gold.total

  return {
    weeklyGold,
    weeklyPurple,
    weeklyBlue,
    weeklyGreen,
    goldSurplus,
    purpleSurplus,
    blueSurplus,
    greenSurplus
  }
})

const doubleWeekRows = computed(() => {
  if (!doubleWeekSummary.value) return []

  const s = doubleWeekSummary.value

  return [
    {
      label: '金',
      weekly: formatNumber(s.weeklyGold),
      surplus: formatNumber(s.goldSurplus),
      remark: `${formatNumber(s.weeklyGold)} + ${formatNumber(s.purpleSurplus)} / 3 - ${roleDemand.gold.total}`
    },
    {
      label: '紫',
      weekly: formatNumber(s.weeklyPurple),
      surplus: formatNumber(s.purpleSurplus),
      remark: `${formatNumber(s.weeklyPurple)} + ${formatNumber(s.blueSurplus)} / 3 - ${roleDemand.purple.total}`
    },
    {
      label: '蓝',
      weekly: formatNumber(s.weeklyBlue),
      surplus: formatNumber(s.blueSurplus),
      remark: `${formatNumber(s.weeklyBlue)} + ${formatNumber(s.greenSurplus)} / 3 - ${roleDemand.blue.total}`
    },
    {
      label: '绿',
      weekly: formatNumber(s.weeklyGreen),
      surplus: formatNumber(s.greenSurplus),
      remark: `${formatNumber(s.weeklyGreen)} - ${roleDemand.green.total}`
    }
  ]
})

const requiredRunsAfterDouble = computed(() => {
  if (!doubleWeekSummary.value || !level8Summary.value || level8Summary.value.combinedGold <= 0) {
    return '0.00'
  }

  const remainingGold = Math.max(0, -doubleWeekSummary.value.goldSurplus)
  return formatNumber(remainingGold / level8Summary.value.combinedGold)
})

const getMaterialRowClass = ({ row }: { row: { label: string } }) => {
  if (row.label === '金') return 'material-row-gold'
  if (row.label === '紫') return 'material-row-purple'
  if (row.label === '蓝') return 'material-row-blue'
  if (row.label === '绿') return 'material-row-green'
  return ''
}

const loadStats = async () => {
  loading.value = true
  try {
    const params: any = {}
    if (filters.player_id) params.player_id = filters.player_id

    const response = await resonanceApi.getDetailedStats(params)
    detailedStats.value = response.data
  } catch (error) {
    ElMessage.error('加载失败: ' + (error as Error).message)
  } finally {
    loading.value = false
  }
}

watch(() => props.refresh, () => {
  loadStats()
})

onMounted(() => {
  loadStats()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.summary-title {
  font-weight: 600;
  margin-bottom: 12px;
}

.no-data {
  text-align: center;
  padding: 40px;
  color: #909399;
  font-size: 14px;
}

:deep(.material-row-gold td) {
  color: #b8860b;
  font-weight: 600;
}

:deep(.material-row-purple td) {
  color: #7d3c98;
  font-weight: 600;
}

:deep(.material-row-blue td) {
  color: #1f78ff;
  font-weight: 600;
}

:deep(.material-row-green td) {
  color: #2e8b57;
  font-weight: 600;
}
</style>
