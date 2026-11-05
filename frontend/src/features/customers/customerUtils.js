import dayjs from 'dayjs'
import buddhistEra from 'dayjs/plugin/buddhistEra'
import 'dayjs/locale/th'

dayjs.extend(buddhistEra)

export const PAGE_SIZE_OPTIONS = [10, 20, 50]

// value '' = every tier (GET /customers without tier)
export const TIER_FILTERS = [
  { value: '', label: 'ทุกระดับ' },
  { value: 'vip', label: 'VIP' },
  { value: 'general', label: 'ทั่วไป' },
]

// value '' = everyone (GET /customers without joined); new/old split at NEW_CUSTOMER_DAYS
export const JOINED_FILTERS = [
  { value: '', label: 'ทั้งหมด' },
  { value: 'new', label: 'ลูกค้าใหม่' },
  { value: 'old', label: 'ลูกค้าเก่า' },
]

export const TIERS = {
  vip: { label: 'VIP', className: 'bg-gold-bg text-gold-ink', avatarClassName: 'bg-gold-bg text-gold-ink' },
  general: { label: 'ทั่วไป', className: 'bg-surface-2 text-muted', avatarClassName: 'bg-surface-2 text-muted' },
}

export const tierOf = (tier) => TIERS[tier] ?? TIERS.general

// "ลูกค้าใหม่" = added within the last 7 × 24 hours (counted from created_at, not calendar days).
// Fixed for now; must match NewCustomerDays in backend/internal/customer/service.go (GET /customers?joined=).
export const NEW_CUSTOMER_DAYS = 7

export const isNewCustomer = (customer) =>
  Boolean(customer.created_at) && Date.now() - new Date(customer.created_at).getTime() < NEW_CUSTOMER_DAYS * 24 * 60 * 60 * 1000

export const fullName = (customer) => [customer.fname, customer.lname].filter(Boolean).join(' ')

export const initialOf = (customer) => customer.fname?.trim().charAt(0) || '?'

export const formatDate = (value) => (value ? dayjs(value).locale('th').format('D MMM BBBB') : '–')

export const formatDateTime = (value) => (value ? dayjs(value).locale('th').format('D MMM BBBB, HH:mm') : '–')

// Date-only columns (b_day, order_date) come back at UTC midnight — keep the YYYY-MM-DD part
// so no timezone can shift the day
export const formatDateOnly = (value) => (value ? formatDate(value.slice(0, 10)) : '–')

// Money comes from the API as decimal strings
export const toNumber = (value) => {
  const number = Number(value)
  return Number.isFinite(number) ? number : 0
}

export const ORDER_STATUSES = {
  paid: { label: 'ชำระครบ', className: 'bg-success-bg text-success' },
  partial: { label: 'ชำระบางส่วน', className: 'bg-warn-bg text-warn' },
  unpaid: { label: 'ยังไม่ชำระ', className: 'bg-danger-bg text-danger' },
}

// Show only the last 4 digits on screen; the full number stays in the edit form
export const maskCitizenId = (value) => (value ? `•••••••••${value.slice(-4)}` : '')
