import dayjs from 'dayjs'
import buddhistEra from 'dayjs/plugin/buddhistEra'
import 'dayjs/locale/th'
import { FiAlertTriangle, FiClock, FiDownload, FiHome, FiSliders } from 'react-icons/fi'

dayjs.extend(buddhistEra)

export const PAGE_SIZE_OPTIONS = [10, 20, 50, 100]

export const DEFAULT_QTY_FILTERS = { search: '', category: '', active: 'true', branchId: '', lowStock: false }
export const DEFAULT_LOG_FILTERS = { preset: '7d', from: '', to: '', stockId: '', logType: '', direction: '', branchId: '' }
export const MAX_PAGE_SIZE = 100 // backend max

export const LOG_TYPES = {
  receive: { label: 'รับเข้า', color: '#1E6B3A' },
  expired: { label: 'หมดอายุ', color: '#7A4E00' },
  damaged: { label: 'เสียหาย', color: '#9B1C1C' },
  internal: { label: 'ใช้ภายใน', color: '#4A4C55' },
  adjust: { label: 'ปรับยอด', color: '#6B4E1E' },
  transfer: { label: 'โอนสาขา', color: '#1F5F6B' },
  sale: { label: 'ขาย', color: '#2F4C8F', auto: true },
  use: { label: 'ใช้ในบริการ', color: '#3F4F62', auto: true },
  open: { label: 'เปิดกระปุก', color: '#5B3566', auto: true },
  count_diff: { label: 'ผลต่างนับสต็อก', color: '#8A3F0C', auto: true },
}

// Types a user can record by hand (POST /stock-logs). sign 0 = user picks + or −.
export const MOVEMENT_TYPES = [
  { value: 'receive', label: 'รับเข้า', hint: 'ของเข้าสต็อก', icon: FiDownload, sign: 1 },
  { value: 'expired', label: 'หมดอายุ', hint: 'ตัดของหมดอายุออก', icon: FiClock, sign: -1 },
  { value: 'damaged', label: 'เสียหาย', hint: 'แตก ชำรุด ใช้ไม่ได้', icon: FiAlertTriangle, sign: -1 },
  { value: 'internal', label: 'ใช้ภายใน', hint: 'ใช้ในร้าน ไม่ได้ขาย', icon: FiHome, sign: -1 },
  { value: 'adjust', label: 'ปรับยอด', hint: 'แก้ยอดที่บันทึกผิด ±', icon: FiSliders, sign: 0 },
]

export const CLOSE_REASONS = {
  finished: { label: 'ใช้หมด', className: 'bg-surface-2 text-muted' },
  damaged: { label: 'เสียหาย', className: 'bg-danger-bg text-danger' },
  expired: { label: 'หมดอายุ', className: 'bg-warn-bg text-warn' },
}

const BRANCH_COLORS = ['#2F4C8F', '#8C2F3D', '#A7B6C6', '#1F5F6B', '#6B4E1E', '#5B3566']

// Colour follows the branch's position in the (code-sorted) branch list, so it never depends on a fixed count
export const branchColor = (branches, branchId) => {
  const index = branches.findIndex((branch) => branch.id === branchId)
  return index < 0 ? '#8A8C95' : BRANCH_COLORS[index % BRANCH_COLORS.length]
}

export const toNumber = (value) => {
  const number = Number(value)
  return Number.isFinite(number) ? number : 0
}

export const isContainer = (stock) => (stock?.uses_per_unit ?? 1) > 1

const th = (value) => dayjs(value).locale('th')
export const formatDate = (value) => th(value).format('D MMM BBBB')
export const formatTime = (value) => th(value).format('HH:mm')
export const formatDateTime = (value) => th(value).format('D MMM BBBB, HH:mm')
export const daysSince = (value) => Math.max(0, dayjs().startOf('day').diff(dayjs(value).startOf('day'), 'day'))

// YYYY-MM-DD in the browser's local day (the backend reads from/to as Thai time)
export const toISODate = (value) => dayjs(value).format('YYYY-MM-DD')

export const DATE_PRESETS = [
  { value: 'today', label: 'วันนี้' },
  { value: '7d', label: '7 วัน' },
  { value: 'month', label: 'เดือนนี้' },
  { value: 'custom', label: 'กำหนดเอง' },
]

export function presetRange(preset) {
  const today = dayjs()
  switch (preset) {
    case 'today':
      return { from: toISODate(today), to: toISODate(today) }
    case '7d':
      return { from: toISODate(today.subtract(6, 'day')), to: toISODate(today) }
    case 'month':
      return { from: toISODate(today.startOf('month')), to: toISODate(today) }
    default:
      return null
  }
}

// Load every page of an endpoint that caps page_size at 100
export async function fetchAllPages(fetcher, params) {
  const all = []
  for (let page = 1; ; page++) {
    const response = await fetcher({ ...params, page, page_size: MAX_PAGE_SIZE })
    all.push(...(response.data ?? []))
    if (page >= (response.meta?.total_pages ?? 1)) return all
  }
}
