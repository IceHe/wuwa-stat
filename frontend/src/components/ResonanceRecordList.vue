<template>
  <el-card>
    <el-form :inline="true" :model="filters" class="filter-form">
      <el-form-item label="玩家">
        <el-input v-model="filters.player_id" placeholder="筛选玩家" clearable style="width: 120px" />
      </el-form-item>
      <el-form-item label="索拉">
        <el-select v-model="filters.sola_level" placeholder="选择等级" clearable style="width: 100px">
          <el-option v-for="level in [8, 7, 6, 5, 4, 3, 2, 1]" :key="level" :label="`等级 ${level}`" :value="level" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleQuery">查询</el-button>
        <el-button @click="loadRecords">刷新</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="records" v-loading="loading" stripe>
      <el-table-column prop="date" label="日期" width="120" />
      <el-table-column prop="player_id" label="玩家ID" width="150" />
      <el-table-column prop="sola_level" label="索拉等级" width="100" />
      <el-table-column prop="claim_count" label="领取次数" width="100" align="center">
        <template #default="{ row }">
          {{ row.claim_count }}次
        </template>
      </el-table-column>
      <el-table-column prop="gold" width="80" align="center">
        <template #header>
          <span class="material-gold">金</span>
        </template>
        <template #default="{ row }">
          <span class="material-gold">{{ row.gold }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="purple" width="80" align="center">
        <template #header>
          <span class="material-purple">紫</span>
        </template>
        <template #default="{ row }">
          <span class="material-purple">{{ row.purple }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="blue" width="80" align="center">
        <template #header>
          <span class="material-blue">蓝</span>
        </template>
        <template #default="{ row }">
          <span class="material-blue">{{ row.blue }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="green" width="80" align="center">
        <template #header>
          <span class="material-green">绿</span>
        </template>
        <template #default="{ row }">
          <span class="material-green">{{ row.green }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="录入时间" width="180">
        <template #default="{ row }">
          {{ formatDateTime(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column v-if="props.canEdit" label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-popconfirm
            v-if="canDeleteRecord(row)"
            title="确定要删除这条记录吗？"
            @confirm="handleDelete(row.id)"
          >
            <template #reference>
              <el-button type="danger" size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-container" v-if="total > 0">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="handlePageChange"
        @size-change="handleSizeChange"
      />
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'
import { resonanceApi, type ResonanceRecord } from '../api'
import { ElMessage } from 'element-plus'

const props = defineProps<{
  refresh: number
  canEdit: boolean
  canManage: boolean
  currentUserId: number | null
}>()

const loading = ref(false)
const records = ref<ResonanceRecord[]>([])
const filters = reactive({
  player_id: '',
  sola_level: undefined as number | undefined
})
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

const loadRecords = async () => {
  loading.value = true
  try {
    const params: any = {
      skip: (currentPage.value - 1) * pageSize.value,
      limit: pageSize.value
    }
    if (filters.player_id) params.player_id = filters.player_id
    if (filters.sola_level) params.sola_level = filters.sola_level

    const response = await resonanceApi.getRecords(params)
    records.value = response.data.data || []
    total.value = response.data.total || 0
  } catch (error) {
    ElMessage.error('加载失败: ' + (error as Error).message)
  } finally {
    loading.value = false
  }
}

const handleQuery = () => {
  currentPage.value = 1
  loadRecords()
}

const handlePageChange = (page: number) => {
  currentPage.value = page
  loadRecords()
}

const handleSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  loadRecords()
}

const formatDateTime = (dateStr: string) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

const canDeleteRecord = (record: ResonanceRecord) => {
  if (props.canManage) return true
  return props.canEdit && props.currentUserId !== null && record.created_by_user_id === props.currentUserId
}

const handleDelete = async (id: number) => {
  try {
    await resonanceApi.deleteRecord(id)
    ElMessage.success('删除成功')
    loadRecords()
  } catch (error) {
    ElMessage.error('删除失败: ' + (error as Error).message)
  }
}

watch(() => props.refresh, () => {
  loadRecords()
})

onMounted(() => {
  loadRecords()
})
</script>

<style scoped>
.filter-form {
  margin-bottom: 0;
}
.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
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
</style>
