<template>
  <el-card>
    <template #header><div class="card-header"><span>声骸主词条统计</span><el-button type="primary" size="small" @click="loadStats">刷新</el-button></div></template>
    <el-form :inline="true" :model="filters">
      <el-form-item label="玩家ID"><el-input v-model="filters.player_id" placeholder="筛选玩家" clearable /></el-form-item>
      <el-form-item label="索拉"><el-select v-model="filters.sola_level" clearable placeholder="选择等级" style="width: 110px"><el-option v-for="level in solaLevels" :key="level" :label="`等级 ${level}`" :value="level" /></el-select></el-form-item>
      <el-form-item><el-button type="primary" @click="loadStats">查询</el-button></el-form-item>
    </el-form>
    <el-tabs v-model="activeView" v-loading="loading">
      <el-tab-pane label="详情统计" name="details">
        <el-table :data="stats.details" stripe border empty-text="暂无数据">
          <el-table-column prop="date" label="日期" width="110" /><el-table-column prop="player_id" label="玩家ID" width="145" /><el-table-column prop="sola_level" label="索拉" width="75" />
          <el-table-column prop="c3_main_stat" label="C3主词条" min-width="120" /><el-table-column prop="c1_main_stat" label="C1主词条" min-width="100" />
          <el-table-column prop="tacet_domain" label="无音区" min-width="220">
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
          <el-table-column prop="count" label="次数" width="75" />
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="汇总统计" name="summary">
        <el-table :data="stats.summary" stripe border empty-text="暂无数据">
          <el-table-column prop="dimension" label="统计项目" width="120" /><el-table-column prop="value" label="项目值" min-width="230">
            <template #default="{ row }">
              <div class="selection-cell">
                <span class="set-icon-list" v-if="getSummaryItems(row.dimension, row.value).length">
                  <img v-for="set in getSummaryItems(row.dimension, row.value)" :key="set.id" :src="set.icon" :alt="set.name" :title="set.name" class="set-icon" />
                </span>
                <span>{{ row.value }}</span>
              </div>
            </template>
          </el-table-column><el-table-column prop="count" label="次数" width="100" /><el-table-column prop="percentage" label="占比" width="100"><template #default="{ row }">{{ row.percentage }}%</template></el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>
  </el-card>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { echoMainApi, type EchoMainStats } from '../api'
import { getEchoSet, getSummaryItems, getTacetDomainSets } from '../data/echoCatalog'
const props = defineProps<{ refresh: number }>(); const loading = ref(false); const activeView = ref('details'); const solaLevels = [8, 7, 6, 5, 4, 3, 2, 1]
const filters = reactive({ player_id: '', sola_level: undefined as number | undefined }); const stats = ref<EchoMainStats>({ details: [], summary: [] })
const loadStats = async () => { loading.value = true; try { stats.value = (await echoMainApi.getStats({ player_id: filters.player_id || undefined, sola_level: filters.sola_level })).data } catch (error) { ElMessage.error('加载失败: ' + (error as Error).message) } finally { loading.value = false } }
watch(() => props.refresh, () => void loadStats()); onMounted(() => void loadStats())
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.selection-cell { display: inline-flex; align-items: center; gap: 6px; min-width: 0; }
.set-icon-list { display: inline-flex; align-items: center; gap: 2px; }
.set-icon { width: 24px; height: 24px; object-fit: contain; flex: 0 0 auto; }
</style>
