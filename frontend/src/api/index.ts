import axios from 'axios'
import { ElMessage } from 'element-plus'

const AUTH_TOKEN_KEY = 'wuwa_auth_token'
const AUTH_UNAUTHORIZED_EVENT = 'wuwa-auth-unauthorized'

let authMeCache: AuthMeResponse | null = null
let authMeCacheToken = ''
let authMePendingRequest: Promise<AuthMeResponse> | null = null
let authMePendingToken = ''

const api = axios.create({
  baseURL: '/api',
  timeout: 10000
})

export type Permission = 'manage' | 'view' | 'edit'

export interface TacetRecord {
  id?: number
  date: string
  player_id: string
  gold_tubes: number
  purple_tubes: number
  claim_count: 1 | 2
  sola_level: number
  created_by_user_id?: number | null
  created_at?: string
}

export interface TacetRecordsResponse {
  data: TacetRecord[]
  total: number
  page_size: number
  current_page: number
}

export interface TacetStats {
  total_records: number
  total_claim_count: number
  total_gold_tubes: number
  total_purple_tubes: number
  avg_gold_tubes: number
  avg_purple_tubes: number
  player_count: number
}

export interface TacetDropCombination {
  gold_tubes: number
  purple_tubes: number
  claim_count: number
  experience: number
  count: number
  percentage: number
}

export interface TacetSolaLevelStats {
  sola_level: number
  combinations: TacetDropCombination[]
  total_count: number
  avg_experience: number
}

export interface TacetDetailedStats {
  level_stats: TacetSolaLevelStats[]
}

export interface AscensionRecord {
  id?: number
  date: string
  player_id: string
  sola_level: number
  drop_count: number
  created_by_user_id?: number | null
  created_at?: string
}

export interface AscensionRecordsResponse {
  data: AscensionRecord[]
  total: number
  page_size: number
  current_page: number
}

export interface AscensionDropCombination {
  drop_count: number
  count: number
  percentage: number
}

export interface AscensionSolaLevelStats {
  sola_level: number
  combinations: AscensionDropCombination[]
  total_count: number
  avg_drop_count: number
}

export interface AscensionDetailedStats {
  level_stats: AscensionSolaLevelStats[]
}

export interface ResonanceRecord {
  id?: number
  date: string
  player_id: string
  sola_level: number
  claim_count: 1 | 2
  gold: number
  purple: number
  blue: number
  green: number
  created_by_user_id?: number | null
  created_at?: string
}

export interface ResonanceRecordsResponse {
  data: ResonanceRecord[]
  total: number
  page_size: number
  current_page: number
}

export interface ResonanceDropCombination {
  claim_count: number
  gold: number
  purple: number
  blue: number
  green: number
  count: number
  percentage: number
}

export interface ResonanceSolaLevelStats {
  sola_level: number
  combinations: ResonanceDropCombination[]
  total_count: number
  total_claim_count: number
  total_gold: number
  total_purple: number
  total_blue: number
  total_green: number
  avg_gold: number
  avg_purple: number
  avg_blue: number
  avg_green: number
}

export interface ResonanceDetailedStats {
  level_stats: ResonanceSolaLevelStats[]
}

export interface AuthMeResponse {
  user_id: number
  name: string
  permissions: Permission[]
}

export const getStoredAuthToken = (): string => localStorage.getItem(AUTH_TOKEN_KEY) || ''

export const setStoredAuthToken = (token: string) => {
  localStorage.setItem(AUTH_TOKEN_KEY, token)
}

export const clearStoredAuthToken = () => {
  localStorage.removeItem(AUTH_TOKEN_KEY)
  authMeCache = null
  authMeCacheToken = ''
  authMePendingRequest = null
  authMePendingToken = ''
}

api.interceptors.request.use((config) => {
  const token = getStoredAuthToken()
  if (token) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error?.response?.status
    if (status === 401) {
      clearStoredAuthToken()
      window.dispatchEvent(new Event(AUTH_UNAUTHORIZED_EVENT))
    } else if (status === 403) {
      ElMessage.error('无权限')
    } else if (status === 503) {
      ElMessage.error('鉴权服务不可用')
    }
    return Promise.reject(error)
  }
)

export const authEvents = {
  unauthorized: AUTH_UNAUTHORIZED_EVENT
}

const fetchAuthMe = async (): Promise<AuthMeResponse> => {
  const response = await api.get<AuthMeResponse>('/auth/me')
  return response.data
}

export const getAuthMe = async (forceRefresh = false): Promise<AuthMeResponse | null> => {
  const token = getStoredAuthToken()
  if (!token) {
    authMeCache = null
    authMeCacheToken = ''
    authMePendingRequest = null
    return null
  }

  const shouldUseCache =
    !forceRefresh &&
    authMeCache !== null &&
    authMeCacheToken === token

  if (shouldUseCache) {
    return authMeCache
  }

  if (!forceRefresh && authMePendingRequest && authMePendingToken === token) {
    return authMePendingRequest
  }

  authMePendingToken = token
  authMePendingRequest = fetchAuthMe()
    .then((authMe) => {
      authMeCache = authMe
      authMeCacheToken = token
      return authMe
    })
    .finally(() => {
      authMePendingRequest = null
      authMePendingToken = ''
    })

  return authMePendingRequest
}

export const authApi = {
  me: () => getAuthMe()
}

export const tacetApi = {
  createRecords: (records: TacetRecord[]) =>
    api.post<TacetRecord[]>('/tacet_records', { tacet_records: records }),

  getRecords: (params?: {
    player_id?: string
    start_date?: string
    end_date?: string
    sola_level?: number
    skip?: number
    limit?: number
  }) => api.get<TacetRecordsResponse>('/tacet_records', { params }),

  getStats: (params?: {
    player_id?: string
    start_date?: string
    end_date?: string
  }) => api.get<TacetStats>('/stats', { params }),

  getDetailedStats: (params?: {
    player_id?: string
    start_date?: string
    end_date?: string
  }) => api.get<TacetDetailedStats>('/detailed-stats', { params }),

  getPlayerIds: () => api.get<string[]>('/player-ids'),

  deleteRecord: (id: number) => api.delete(`/tacet_records/${id}`)
}

export const ascensionApi = {
  createRecords: (records: AscensionRecord[]) =>
    api.post<AscensionRecord[]>('/ascension-records', { ascension_records: records }),

  getRecords: (params?: {
    player_id?: string
    start_date?: string
    end_date?: string
    sola_level?: number
    skip?: number
    limit?: number
  }) => api.get<AscensionRecordsResponse>('/ascension-records', { params }),

  getDetailedStats: (params?: {
    player_id?: string
    start_date?: string
    end_date?: string
  }) => api.get<AscensionDetailedStats>('/ascension-detailed-stats', { params }),

  getPlayerIds: () => api.get<string[]>('/ascension-player-ids'),

  deleteRecord: (id: number) => api.delete(`/ascension-records/${id}`)
}

export const resonanceApi = {
  createRecords: (records: ResonanceRecord[]) =>
    api.post<ResonanceRecord[]>('/resonance-records', { resonance_records: records }),

  getRecords: (params?: {
    player_id?: string
    start_date?: string
    end_date?: string
    sola_level?: number
    skip?: number
    limit?: number
  }) => api.get<ResonanceRecordsResponse>('/resonance-records', { params }),

  getDetailedStats: (params?: {
    player_id?: string
    start_date?: string
    end_date?: string
  }) => api.get<ResonanceDetailedStats>('/resonance-detailed-stats', { params }),

  getPlayerIds: () => api.get<string[]>('/resonance-player-ids'),

  deleteRecord: (id: number) => api.delete(`/resonance-records/${id}`)
}
