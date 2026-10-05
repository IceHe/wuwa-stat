<template>
  <el-card>
    <template #header><span>声骸主词条录入详情</span></template>
    <el-form :inline="true" :model="filters" class="filter-form">
      <el-form-item label="玩家"><el-input v-model="filters.player_id" placeholder="筛选玩家" clearable style="width: 130px" /></el-form-item>
      <el-form-item label="索拉"><el-select v-model="filters.sola_level" clearable placeholder="选择等级" style="width: 110px"><el-option v-for="level in solaLevels" :key="level" :label="`等级 ${level}`" :value="level" /></el-select></el-form-item>
      <el-form-item><el-button type="primary" @click="query">查询</el-button><el-button @click="loadRecords">刷新</el-button></el-form-item>
    </el-form>
    <el-table :data="records" v-loading="loading" stripe>
      <el-table-column prop="date" label="日期" width="110" />
      <el-table-column prop="player_id" label="玩家ID" width="145" />
      <el-table-column prop="sola_level" label="索拉等级" width="95" />
      <el-table-column prop="c3_main_stat" label="C3主词条" min-width="120" />
      <el-table-column prop="c1_main_stat" label="C1主词条" min-width="100" />
      <el-table-column prop="tacet_domain" label="所属无音区" min-width="220">
        <template #default="{ row }">
          <div class="selection-cell">
            <span>{{ row.tacet_domain }}</span>
            <span v-if="getTacetDomainSets(row.tacet_domain).length" class="set-icon-list">
              <img v-for="set in getTacetDomainSets(row.tacet_domain)" :key="set.id" :src="set.icon" :alt="set.name" :title="set.name" class="set-icon" />
            </span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="echo_set" label="声骸套装" min-width="190">
        <template #default="{ row }">
          <div class="selection-cell">
            <img v-if="getEchoSet(row.echo_set)" :src="getEchoSet(row.echo_set)?.icon" :alt="row.echo_set" :title="row.echo_set" class="set-icon" />
            <span>{{ row.echo_set }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="录入时间" width="180"><template #default="{ row }">{{ formatDateTime(row.created_at) }}</template></el-table-column>
      <el-table-column v-if="props.canEdit" label="操作" width="100" fixed="right"><template #default="{ row }"><el-popconfirm v-if="canDeleteRecord(row)" title="确定要删除这条记录吗？" @confirm="handleDelete(row.id!)"><template #reference><el-button type="danger" size="small">删除</el-button></template></el-popconfirm></template></el-table-column>
    </el-table>
    <div v-if="total > 0" class="pagination-container"><el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize" :page-sizes="[10, 20, 50, 100]" :total="total" layout="total, sizes, prev, pager, next, jumper" @current-change="loadRecords" @size-change="handleSizeChange" /></div>
  </el-card>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { echoMainApi, type EchoMainRecord } from '../api'
import { getEchoSet, getTacetDomainSets } from '../data/echoCatalog'
const props = defineProps<{ refresh: number; canEdit: boolean; canManage: boolean; currentUserId: number | null }>()
const records = ref<EchoMainRecord[]>([]); const loading = ref(false); const total = ref(0); const currentPage = ref(1); const pageSize = ref(20)
const filters = reactive({ player_id: '', sola_level: undefined as number | undefined }); const solaLevels = [8, 7, 6, 5, 4, 3, 2, 1]
const loadRecords = async () => { loading.value = true; try { const response = await echoMainApi.getRecords({ skip: (currentPage.value - 1) * pageSize.value, limit: pageSize.value, player_id: filters.player_id || undefined, sola_level: filters.sola_level }); records.value = response.data.data || []; total.value = response.data.total || 0 } catch (error) { ElMessage.error('加载失败: ' + (error as Error).message) } finally { loading.value = false } }
const query = () => { currentPage.value = 1; void loadRecords() }; const handleSizeChange = (size: number) => { pageSize.value = size; currentPage.value = 1; void loadRecords() }
const formatDateTime = (value?: string) => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : ''
const canDeleteRecord = (record: EchoMainRecord) => props.canManage || (props.canEdit && props.currentUserId !== null && record.created_by_user_id === props.currentUserId)
const handleDelete = async (id: number) => { try { await echoMainApi.deleteRecord(id); ElMessage.success('删除成功'); void loadRecords() } catch (error) { ElMessage.error('删除失败: ' + (error as Error).message) } }
watch(() => props.refresh, () => void loadRecords()); onMounted(() => void loadRecords())
</script>

<style scoped>
.filter-form { margin-bottom: 0; }.pagination-container { display: flex; justify-content: flex-end; margin-top: 12px; }
.selection-cell { display: inline-flex; align-items: center; gap: 6px; min-width: 0; }
.set-icon-list { display: inline-flex; align-items: center; gap: 2px; }
.set-icon { width: 24px; height: 24px; object-fit: contain; flex: 0 0 auto; }
</style>
