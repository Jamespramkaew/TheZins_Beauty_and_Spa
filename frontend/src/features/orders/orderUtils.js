import { roundMoney, toNumber } from '../../utils/number'

export const PAGE_SIZE_OPTIONS = [10, 20, 50]

// Methods a cashier can pick (credit = VIP wallet, recorded by the system only)
export const PAYMENT_METHODS = [
  { value: 'cash', label: 'เงินสด' },
  { value: 'transfer', label: 'โอนเงิน' },
  { value: 'card', label: 'บัตร' },
]

const METHOD_LABELS = { cash: 'เงินสด', transfer: 'โอนเงิน', card: 'บัตร', credit: 'เครดิต VIP' }
export const methodLabel = (method) => METHOD_LABELS[method] ?? method

// Same rule as the backend for money: not negative, at most 2 decimals
export const MONEY_PATTERN = /^\d+(\.\d{1,2})?$/

// order.source (optional) — '' = not asked
export const SOURCES = [
  { value: '', label: 'ไม่ระบุ' },
  { value: 'walk-in', label: 'walk-in' },
  { value: 'phone', label: 'โทรศัพท์' },
  { value: 'page', label: 'เพจ' },
]

const SOURCE_LABELS = { phone: 'โทรศัพท์', 'walk-in': 'walk-in', page: 'เพจ' }
export const sourceLabel = (source) => (source ? (SOURCE_LABELS[source] ?? source) : '')

// One line of a bill: qty × unit_price − discount (same formula as the backend total)
export const lineTotal = (item) => roundMoney(item.qty * toNumber(item.unit_price) - toNumber(item.discount))

export const itemName = (item) => item.course?.name ?? item.stock?.name ?? (item.item_type === 'credit' ? 'เครดิต VIP' : '–')

const ITEM_TYPE_LABELS = { course: 'คอร์ส', stock: 'สินค้า', credit: 'เครดิต' }
export const itemTypeLabel = (type) => ITEM_TYPE_LABELS[type] ?? type

// value '' = every status (GET /orders without status)
export const STATUS_FILTERS = [
  { value: '', label: 'ทั้งหมด' },
  { value: 'unpaid', label: 'ยังไม่ชำระ' },
  { value: 'partial', label: 'ชำระบางส่วน' },
  { value: 'paid', label: 'ชำระครบ' },
]
