export interface EchoSetOption {
  id: number
  name: string
  icon: string
}

export interface TacetDomainOption {
  name: string
  setIds: number[]
}

const icon = (id: number) => `/echo-icons/sonata-${id}.webp`

export const c3MainStats = [
  '攻击',
  '生命',
  '防御',
  '冷凝伤害',
  '热熔伤害',
  '气动伤害',
  '导电伤害',
  '衍射伤害',
  '湮灭伤害',
  '共鸣效率'
] as const

export const c1MainStats = ['攻击', '生命', '防御'] as const

// Names and IDs are sourced from nanoka.cc's current sonata.json data.
export const echoSets: EchoSetOption[] = [
  { id: 1, name: '凝夜白霜', icon: icon(1) },
  { id: 2, name: '熔山裂谷', icon: icon(2) },
  { id: 3, name: '彻空冥雷', icon: icon(3) },
  { id: 4, name: '啸谷长风', icon: icon(4) },
  { id: 5, name: '浮星祛暗', icon: icon(5) },
  { id: 6, name: '沉日劫明', icon: icon(6) },
  { id: 7, name: '隐世回光', icon: icon(7) },
  { id: 8, name: '轻云出月', icon: icon(8) },
  { id: 9, name: '不绝余音', icon: icon(9) },
  { id: 10, name: '凌冽决断之心', icon: icon(10) },
  { id: 11, name: '此间永驻之光', icon: icon(11) },
  { id: 12, name: '幽夜隐匿之帷', icon: icon(12) },
  { id: 13, name: '高天共奏之曲', icon: icon(13) },
  { id: 14, name: '无惧浪涛之勇', icon: icon(14) },
  { id: 16, name: '流云逝尽之空', icon: icon(16) },
  { id: 17, name: '愿戴荣光之旅', icon: icon(17) },
  { id: 18, name: '奔狼燎原之焰', icon: icon(18) },
  { id: 19, name: '失序彼岸之梦', icon: icon(19) },
  { id: 20, name: '荣斗铸锋之冠', icon: icon(20) },
  { id: 21, name: '息界同调之律', icon: icon(21) },
  { id: 22, name: '焚羽猎魔之影', icon: icon(22) },
  { id: 23, name: '命理崩毁之弦', icon: icon(23) },
  { id: 24, name: '逆光跃彩之约', icon: icon(24) },
  { id: 25, name: '星构寻辉之环', icon: icon(25) },
  { id: 26, name: '流金溯真之式', icon: icon(26) },
  { id: 27, name: '长路启航之星', icon: icon(27) },
  { id: 28, name: '斑驳粉饰之沫', icon: icon(28) },
  { id: 29, name: '听唤语义之愿', icon: icon(29) },
  { id: 30, name: '雪落无声之愿', icon: icon(30) },
  { id: 31, name: '剪心辑梦之影', icon: icon(31) },
  { id: 32, name: '碎梦亡鬼之魇', icon: icon(32) },
  { id: 33, name: '羽落空尘之歌', icon: icon(33) },
  { id: 34, name: '清邪荡煞之心', icon: icon(34) },
  { id: 35, name: '冥途夜行之灯', icon: icon(35) },
  { id: 36, name: '衔梦照世之心', icon: icon(36) },
  { id: 37, name: '镜影流电之瞬', icon: icon(37) },
  { id: 38, name: '茜染怀想之花', icon: icon(38) }
]

// Field names and reward pairs are cross-checked against Game8's current
// Tacet Field Locations and Rewards page. The page groups the fields by
// Mengzhou, Lahai-Roi, Rinascita, and Huanglong; each selectable item below
// is one actual field rather than an aggregate region.
export const tacetDomains: TacetDomainOption[] = [
  // Mengzhou
  { name: '无音区·Heart of Flames', setIds: [36, 37] },
  { name: '无音区·Heart of Stillness', setIds: [36, 38] },
  { name: '无音区·Eastern Xuan Peaks', setIds: [33, 34] },
  { name: '无音区·Western Fang Peaks', setIds: [33, 35] },
  // Lahai-Roi
  { name: '无音区·Solisia Landing', setIds: [30, 31] },
  { name: '无音区·Mount Gjallar', setIds: [27, 28] },
  { name: '无音区·Frostlands Transit Port', setIds: [27, 29] },
  { name: '无音区·Mawburrow Desert', setIds: [26, 24] },
  { name: '无音区·Stagnant Run', setIds: [26, 25] },
  // Rinascita / 黎那汐塔
  { name: '无音区·Mournfell Canyon', setIds: [17, 18] },
  { name: '无音区·Beohr Waters', setIds: [16, 12] },
  { name: '无音区·Riccioli Islands', setIds: [11, 14] },
  { name: '无音区·Fagaceae Peninsula', setIds: [10, 11] },
  { name: "无音区·Penitent's End", setIds: [12, 13] },
  // Huanglong
  { name: '无音区·Central Plains', setIds: [9, 7] },
  { name: '无音区·Desorock Highland', setIds: [4, 1] },
  { name: '无音区·Misty Coast', setIds: [3, 2] },
  { name: "无音区·Tiger's Maw Mine / Wuming Bay", setIds: [1, 3] },
  { name: '无音区·Port City of Guixu', setIds: [2, 7] },
  { name: '无音区·Dim Forest', setIds: [6, 8] },
  { name: "无音区·Whining Aix's Mire", setIds: [5, 9] }
]

const echoSetByName = new Map(echoSets.map((set) => [set.name, set]))

// These aliases keep icons visible for records created with the old labels.
const echoSetAliases: Record<string, string> = {
  浮星祈愿: '浮星祛暗',
  失序彼岸: '失序彼岸之梦',
  此间永驻: '此间永驻之光',
  愿戴荣光: '愿戴荣光之旅',
  无惧浪涛: '无惧浪涛之勇'
}

export const getEchoSet = (name: string | undefined): EchoSetOption | undefined => {
  if (!name) return undefined
  return echoSetByName.get(name) || echoSetByName.get(echoSetAliases[name])
}

export const getTacetDomain = (name: string | undefined): TacetDomainOption | undefined => {
  if (!name) return undefined
  return tacetDomains.find((domain) => domain.name === name)
}

export const getTacetDomainSets = (name: string | undefined): EchoSetOption[] => {
  const domain = getTacetDomain(name)
  if (!domain) return []
  const byId = new Map(echoSets.map((set) => [set.id, set]))
  return domain.setIds.map((id) => byId.get(id)).filter((set): set is EchoSetOption => set !== undefined)
}

export const getSummaryItems = (dimension: string, value: string) => {
  if (dimension === '声骸套装') {
    const set = getEchoSet(value)
    return set ? [set] : []
  }
  if (dimension === '无音区') return getTacetDomainSets(value)
  return []
}
