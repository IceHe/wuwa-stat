<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span>突破材料掉落统计</span>
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
      <div v-if="detailedStats.level_stats.length === 0" class="no-data">
        暂无数据
      </div>
      <div v-else>
        <el-table
          v-for="levelStat in detailedStats.level_stats"
          :key="levelStat.sola_level"
          :data="levelStat.combinations"
          border
          style="margin-bottom: 20px"
        >
          <el-table-column label="索拉等级" width="100" align="center">
            <template #default="{ $index }">
              <span v-if="$index === 0">{{ levelStat.sola_level }}</span>
            </template>
          </el-table-column>
          <el-table-column label="掉落数量" width="120" align="center">
            <template #default="{ row }">
              {{ row.drop_count }}
            </template>
          </el-table-column>
          <el-table-column label="次数" width="100" align="center">
            <template #default="{ row }">
              {{ row.count }}
            </template>
          </el-table-column>
          <el-table-column label="占比" width="100" align="center">
            <template #default="{ row }">
              {{ row.percentage }}%
            </template>
          </el-table-column>
          <el-table-column label="总次数" width="100" align="center">
            <template #default="{ $index }">
              <span v-if="$index === 0">{{ levelStat.total_count }}</span>
            </template>
          </el-table-column>
          <el-table-column label="平均掉落" width="120" align="center">
            <template #default="{ $index }">
              <span v-if="$index === 0">{{ levelStat.avg_drop_count }}</span>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'
import { ascensionApi, type AscensionDetailedStats } from '../api'
import { ElMessage } from 'element-plus'

const props = defineProps<{
  refresh: number
}>()

const loading = ref(false)
const detailedStats = ref<AscensionDetailedStats>({
  level_stats: []
})

const filters = reactive({
  player_id: ''
})

const loadStats = async () => {
  loading.value = true
  try {
    const params: any = {}
    if (filters.player_id) params.player_id = filters.player_id

    const response = await ascensionApi.getDetailedStats(params)
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

.no-data {
  text-align: center;
  padding: 40px;
  color: #909399;
  font-size: 14px;
}
</style>
