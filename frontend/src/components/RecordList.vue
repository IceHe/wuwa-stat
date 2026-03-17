<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span>产出记录列表</span>
        <el-button type="primary" size="small" @click="loadRecords">
          刷新
        </el-button>
      </div>
    </template>

    <el-form :inline="true" :model="filters">
      <el-form-item label="玩家ID">
        <el-input v-model="filters.player_id" placeholder="筛选玩家" clearable />
      </el-form-item>
      <el-form-item label="索拉等级">
        <el-select v-model="filters.sola_level" placeholder="选择等级" clearable>
          <el-option v-for="level in [6, 7, 8, 9, 10]" :key="level" :label="`等级 ${level}`" :value="level" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="loadRecords">查询</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="records" v-loading="loading" stripe>
      <el-table-column prop="date" label="日期" width="120" />
      <el-table-column prop="player_id" label="玩家ID" width="150" />
      <el-table-column prop="sola_level" label="索拉等级" width="100" />
      <el-table-column prop="gold_tubes" label="金色密音筒" width="120" />
      <el-table-column prop="purple_tubes" label="紫色密音筒" width="120" />
      <el-table-column prop="created_at" label="录入时间" />
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'
import { recordApi, type Record } from '../api'
import { ElMessage } from 'element-plus'

const props = defineProps<{
  refresh: number
}>()

const loading = ref(false)
const records = ref<Record[]>([])
const filters = reactive({
  player_id: '',
  sola_level: undefined as number | undefined
})

const loadRecords = async () => {
  loading.value = true
  try {
    const params: any = { limit: 100 }
    if (filters.player_id) params.player_id = filters.player_id
    if (filters.sola_level) params.sola_level = filters.sola_level

    const response = await recordApi.getRecords(params)
    records.value = response.data
  } catch (error) {
    ElMessage.error('加载失败: ' + (error as Error).message)
  } finally {
    loading.value = false
  }
}

watch(() => props.refresh, () => {
  loadRecords()
})

onMounted(() => {
  loadRecords()
})
</script>
