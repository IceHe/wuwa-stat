<template>
  <el-card>
    <template #header><span>录入无音区声骸主词条</span></template>
    <el-form :model="form" label-width="120px">
      <el-form-item label="统计日期">
        <el-date-picker v-model="form.date" type="date" value-format="YYYY-MM-DD" style="width: 100%" @change="dateManuallyEdited = true" />
      </el-form-item>
      <el-form-item>
        <template #label>
          <button type="button" class="player-id-label-button" @click="openPlayerIdDialog">玩家ID</button>
        </template>
        <PlayerIdField ref="playerIdFieldRef" v-model="form.player_id" :player-ids="playerIds" />
      </el-form-item>
      <el-form-item label="索拉等级">
        <div class="option-button-group">
          <el-button v-for="level in solaLevels" :key="level" :type="form.sola_level === level ? 'primary' : 'default'" @click="form.sola_level = level">等级 {{ level }}</el-button>
        </div>
      </el-form-item>
      <el-form-item label="C3主词条">
        <div class="wide-option-button-group">
          <el-button v-for="stat in c3MainStats" :key="stat" :type="form.c3_main_stat === stat ? 'primary' : 'default'" @click="form.c3_main_stat = stat">{{ stat }}</el-button>
        </div>
      </el-form-item>
      <el-form-item label="C1主词条">
        <div class="wide-option-button-group">
          <el-button v-for="stat in c1MainStats" :key="stat" :type="form.c1_main_stat === stat ? 'primary' : 'default'" @click="form.c1_main_stat = stat">{{ stat }}</el-button>
        </div>
      </el-form-item>
      <el-form-item label="所属无音区">
        <div class="wide-option-button-group">
          <el-button v-for="domain in tacetDomains" :key="domain.name" class="domain-option-button" :type="form.tacet_domain === domain.name ? 'primary' : 'default'" @click="selectDomain(domain.name)">
            <span>{{ domain.name }}</span>
            <span v-if="getTacetDomainSets(domain.name).length" class="domain-set-icons" aria-label="包含的声骸套装">
              <img v-for="set in getTacetDomainSets(domain.name)" :key="set.id" :src="set.icon" :alt="set.name" :title="set.name" class="set-icon set-icon-small" />
            </span>
          </el-button>
          <el-button :type="customDomainSelected ? 'primary' : 'default'" @click="customDomainSelected = true">其他</el-button>
        </div>
        <el-input v-if="customDomainSelected" v-model="form.tacet_domain" class="custom-option-input" placeholder="输入无音区名称" clearable />
      </el-form-item>
      <el-form-item label="声骸套装">
        <div class="wide-option-button-group">
          <el-button v-for="echoSet in echoSets" :key="echoSet.id" class="echo-set-option-button" :type="form.echo_set === echoSet.name ? 'primary' : 'default'" @click="selectEchoSet(echoSet.name)">
            <img :src="echoSet.icon" :alt="echoSet.name" :title="echoSet.name" class="set-icon" />
            <span>{{ echoSet.name }}</span>
          </el-button>
          <el-button :type="customEchoSetSelected ? 'primary' : 'default'" @click="customEchoSetSelected = true">其他</el-button>
        </div>
        <el-input v-if="customEchoSetSelected" v-model="form.echo_set" class="custom-option-input" placeholder="输入声骸套装名称" clearable />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="loading" @click="handleSubmit">提交</el-button>
        <el-button @click="resetForm">重置</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { echoMainApi } from '../api'
import { confirmRecordWithoutEnergyDeduction, isEnergyInsufficientError } from '../utils/energyDeduction'
import { c1MainStats, c3MainStats, echoSets, getTacetDomainSets, tacetDomains } from '../data/echoCatalog'
import PlayerIdField from './PlayerIdField.vue'

const props = withDefaults(defineProps<{ playerId?: string }>(), { playerId: '' })
const emit = defineEmits<{ (e: 'success'): void; (e: 'update:playerId', value: string): void }>()
type PlayerIdFieldExpose = { openDialog: () => void; refreshAccount: () => Promise<void> }
const playerIdFieldRef = ref<PlayerIdFieldExpose | null>(null)
const playerIds = ref<string[]>([])
const loading = ref(false)
const dateManuallyEdited = ref(false)
let dateTimer: ReturnType<typeof setTimeout> | null = null

const solaLevels = [8, 7, 6, 5, 4, 3, 2, 1]
const customDomainSelected = ref(false)
const customEchoSetSelected = ref(false)

const formatDate = (date: Date) => `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
const defaultDate = () => {
  const date = new Date()
  if (date.getHours() < 4) date.setDate(date.getDate() - 1)
  return formatDate(date)
}
const nextDateSwitch = () => {
  const date = new Date()
  date.setHours(4, 0, 0, 0)
  if (Date.now() >= date.getTime()) date.setDate(date.getDate() + 1)
  return date
}
const scheduleDate = () => {
  if (dateTimer) clearTimeout(dateTimer)
  dateTimer = setTimeout(() => {
    if (!dateManuallyEdited.value) form.date = defaultDate()
    scheduleDate()
  }, Math.max(nextDateSwitch().getTime() - Date.now() + 1000, 1000))
}

const form = reactive({
  date: defaultDate(),
  player_id: props.playerId,
  sola_level: 8,
  c3_main_stat: c3MainStats[0] as string,
  c1_main_stat: c1MainStats[0] as string,
  tacet_domain: '',
  echo_set: ''
})

const selectDomain = (domain: string) => { customDomainSelected.value = false; form.tacet_domain = domain }
const selectEchoSet = (echoSet: string) => { customEchoSetSelected.value = false; form.echo_set = echoSet }

const resetForm = () => {
  form.date = defaultDate()
  form.player_id = props.playerId
  form.sola_level = 8
  form.c3_main_stat = c3MainStats[0]
  form.c1_main_stat = c1MainStats[0]
  form.tacet_domain = ''
  form.echo_set = ''
  customDomainSelected.value = false
  customEchoSetSelected.value = false
  dateManuallyEdited.value = false
}
const openPlayerIdDialog = () => playerIdFieldRef.value?.openDialog()

watch(() => props.playerId, value => { if (form.player_id !== value) form.player_id = value })
watch(() => form.player_id, value => { if (value !== props.playerId) emit('update:playerId', value) })

const loadPlayerIds = async () => {
  try { playerIds.value = (await echoMainApi.getPlayerIds()).data } catch (error) { console.error('加载玩家ID列表失败:', error) }
}
const submitRecord = async (skipEnergyDeduction = false) => {
  await echoMainApi.createRecords([form], { skipEnergyDeduction })
  void playerIdFieldRef.value?.refreshAccount()
  ElMessage.success('录入成功')
  emit('success')
  resetForm()
}
const handleSubmit = async () => {
  if (!form.player_id) { ElMessage.warning('请输入玩家ID'); return }
  if (!form.tacet_domain || !form.echo_set) { ElMessage.warning('请选择或输入无音区和声骸套装'); return }
  loading.value = true
  try {
    await submitRecord()
  } catch (error) {
    if (!isEnergyInsufficientError(error)) {
      ElMessage.error('录入失败: ' + (error as Error).message)
      loading.value = false
      return
    }
    loading.value = false
    if (!await confirmRecordWithoutEnergyDeduction()) return
    loading.value = true
    try { await submitRecord(true) } catch (retryError) { ElMessage.error('录入失败: ' + (retryError as Error).message) }
  } finally { loading.value = false }
}

onMounted(() => { loadPlayerIds(); scheduleDate() })
onBeforeUnmount(() => { if (dateTimer) clearTimeout(dateTimer) })
</script>

<style scoped>
.player-id-label-button { appearance: none; background: transparent; border: 0; color: inherit; cursor: pointer; font: inherit; padding: 0; }
.player-id-label-button:hover, .player-id-label-button:focus-visible { color: #409eff; }
.wide-option-button-group { display: flex; flex-wrap: wrap; gap: 8px; width: 100%; }
.wide-option-button-group .el-button { margin-left: 0; margin-right: 0; }
.custom-option-input { margin-top: 8px; max-width: 360px; }
.domain-option-button, .echo-set-option-button { display: inline-flex; align-items: center; gap: 6px; min-height: 36px; }
.domain-set-icons { display: inline-flex; align-items: center; gap: 2px; }
.set-icon { width: 24px; height: 24px; object-fit: contain; flex: 0 0 auto; }
.set-icon-small { width: 20px; height: 20px; }
</style>
