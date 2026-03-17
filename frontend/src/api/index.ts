import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000
})

export interface Record {
  id?: number
  date: string
  player_id: string
  gold_tubes: number
  purple_tubes: number
  sola_level: number
  created_at?: string
}

export interface Stats {
  total_records: number
  total_gold_tubes: number
  total_purple_tubes: number
  avg_gold_tubes: number
  avg_purple_tubes: number
  player_count: number
}

export interface DropCombination {
  gold_tubes: number
  purple_tubes: number
  experience: number
  count: number
  percentage: number
}

export interface SolaLevelStats {
  sola_level: number
  combinations: DropCombination[]
  total_count: number
  avg_experience: number
}

export interface DetailedStats {
  level_stats: SolaLevelStats[]
}

export const recordApi = {
  createRecords: (records: Record[]) =>
    api.post<Record[]>('/records', { records }),

  getRecords: (params?: {
    player_id?: string
    start_date?: string
    end_date?: string
    sola_level?: number
    skip?: number
    limit?: number
  }) => api.get<Record[]>('/records', { params }),

  getStats: (params?: {
    player_id?: string
    start_date?: string
    end_date?: string
  }) => api.get<Stats>('/stats', { params }),

  getDetailedStats: (params?: {
    player_id?: string
    start_date?: string
    end_date?: string
  }) => api.get<DetailedStats>('/detailed-stats', { params }),

  getPlayerIds: () => api.get<string[]>('/player-ids'),

  deleteRecord: (id: number) => api.delete(`/records/${id}`)
}
