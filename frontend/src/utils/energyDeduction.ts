import { ElMessageBox } from 'element-plus'

const ENERGY_INSUFFICIENT_CODE = 'energy_insufficient'
const ENERGY_INSUFFICIENT_MESSAGE = '体力不足以扣除，是否只记录不扣体力？'

type ErrorWithResponse = {
  response?: {
    data?: {
      code?: string
      detail?: string
    }
  }
}

export const isEnergyInsufficientError = (error: unknown) => {
  const data = (error as ErrorWithResponse)?.response?.data
  return data?.code === ENERGY_INSUFFICIENT_CODE || data?.detail === ENERGY_INSUFFICIENT_MESSAGE
}

export const confirmRecordWithoutEnergyDeduction = async () => {
  try {
    await ElMessageBox.confirm(ENERGY_INSUFFICIENT_MESSAGE, '体力不足', {
      confirmButtonText: '是',
      cancelButtonText: '否',
      type: 'warning'
    })
    return true
  } catch {
    return false
  }
}
