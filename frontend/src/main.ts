import { createApp } from 'vue'
import {
  ElAlert,
  ElAutocomplete,
  ElButton,
  ElCard,
  ElContainer,
  ElDatePicker,
  ElDescriptions,
  ElDescriptionsItem,
  ElDialog,
  ElForm,
  ElFormItem,
  ElHeader,
  ElInput,
  ElInputNumber,
  ElLoading,
  ElMain,
  ElOption,
  ElPagination,
  ElPopconfirm,
  ElResult,
  ElSelect,
  ElTabPane,
  ElTable,
  ElTableColumn,
  ElTabs,
  ElText,
  ElTooltip
} from 'element-plus'
import 'element-plus/theme-chalk/el-alert.css'
import 'element-plus/theme-chalk/base.css'
import 'element-plus/theme-chalk/el-autocomplete.css'
import 'element-plus/theme-chalk/el-button.css'
import 'element-plus/theme-chalk/el-card.css'
import 'element-plus/theme-chalk/el-container.css'
import 'element-plus/theme-chalk/el-date-picker-panel.css'
import 'element-plus/theme-chalk/el-descriptions.css'
import 'element-plus/theme-chalk/el-descriptions-item.css'
import 'element-plus/theme-chalk/el-dialog.css'
import 'element-plus/theme-chalk/el-form.css'
import 'element-plus/theme-chalk/el-form-item.css'
import 'element-plus/theme-chalk/el-header.css'
import 'element-plus/theme-chalk/el-icon.css'
import 'element-plus/theme-chalk/el-input.css'
import 'element-plus/theme-chalk/el-input-number.css'
import 'element-plus/theme-chalk/el-loading.css'
import 'element-plus/theme-chalk/el-main.css'
import 'element-plus/theme-chalk/el-message.css'
import 'element-plus/theme-chalk/el-message-box.css'
import 'element-plus/theme-chalk/el-option.css'
import 'element-plus/theme-chalk/el-overlay.css'
import 'element-plus/theme-chalk/el-pagination.css'
import 'element-plus/theme-chalk/el-popconfirm.css'
import 'element-plus/theme-chalk/el-popover.css'
import 'element-plus/theme-chalk/el-popper.css'
import 'element-plus/theme-chalk/el-result.css'
import 'element-plus/theme-chalk/el-scrollbar.css'
import 'element-plus/theme-chalk/el-select.css'
import 'element-plus/theme-chalk/el-tab-pane.css'
import 'element-plus/theme-chalk/el-table.css'
import 'element-plus/theme-chalk/el-table-column.css'
import 'element-plus/theme-chalk/el-tabs.css'
import 'element-plus/theme-chalk/el-text.css'
import 'element-plus/theme-chalk/el-time-picker.css'
import 'element-plus/theme-chalk/el-tooltip.css'
import App from './App.vue'

const app = createApp(App)

const components = [
  ElAlert,
  ElAutocomplete,
  ElButton,
  ElCard,
  ElContainer,
  ElDatePicker,
  ElDescriptions,
  ElDescriptionsItem,
  ElDialog,
  ElForm,
  ElFormItem,
  ElHeader,
  ElInput,
  ElInputNumber,
  ElMain,
  ElOption,
  ElPagination,
  ElPopconfirm,
  ElResult,
  ElSelect,
  ElTabPane,
  ElTable,
  ElTableColumn,
  ElTabs,
  ElText,
  ElTooltip
]

components.forEach((component) => {
  app.use(component)
})

app.use(ElLoading)
app.mount('#app')
