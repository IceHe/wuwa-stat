<template>
  <el-container class="container">
    <el-header>
      <h1>鸣潮无音区产出统计</h1>
    </el-header>

    <el-main>
      <el-tabs v-model="activeTab">
        <el-tab-pane label="录入数据" name="input">
          <RecordInput @success="handleInputSuccess" />
        </el-tab-pane>

        <el-tab-pane label="查看记录" name="records">
          <RecordList :refresh="refreshTrigger" />
        </el-tab-pane>

        <el-tab-pane label="统计数据" name="stats">
          <StatsView :refresh="refreshTrigger" />
        </el-tab-pane>
      </el-tabs>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import RecordInput from './components/RecordInput.vue'
import RecordList from './components/RecordList.vue'
import StatsView from './components/StatsView.vue'

const activeTab = ref('input')
const refreshTrigger = ref(0)

const handleInputSuccess = () => {
  refreshTrigger.value++
  activeTab.value = 'records'
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background: #f5f7fa;
}

el-header {
  background: #409eff;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
}

h1 {
  margin: 0;
  font-size: 24px;
}
</style>
