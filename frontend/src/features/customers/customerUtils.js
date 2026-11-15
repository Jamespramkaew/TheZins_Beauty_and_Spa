import dayjs from 'dayjs'

export const PAGE_SIZE_OPTIONS = [10, 20, 50]

// One row of filter buttons. Exactly one is active; `params` is what it adds to GET /customers.
// "ลูกค้าใหม่" covers every tier (added within NEW_CUSTOMER_DAYS).
export const CUSTOMER_FILTERS = [
  { value: '', label: 'ทุกระดับ', params: {} },
  { value: 'vip', label: 'VIP', params: { tier: 'vip' } },
  { value: 'general', label: 'ทั่วไป', params: { tier: 'general' } },
  { value: 'credit', label: 'วงเงิน', params: { tier: 'credit' } },
  { value: 'new', label: 'ลูกค้าใหม่', params: { joined: 'new' } },
]

export const TIERS = {
  vip: { label: 'VIP', className: 'bg-gold/50 text-gold-ink', avatarClassName: 'bg-gold/30 text-gold-ink' },
  general: { label: 'ทั่วไป', className: 'bg-surface-2 text-muted', avatarClassName: 'bg-surface-2 text-muted' },
  credit: { label: '$ วงเงิน', className: 'bg-chart/80 text-white', avatarClassName: 'bg-chart/80 text-white' },
}

// Every tier the backend accepts (validTiers in backend/internal/customer/service.go), in display order
export const TIER_VALUES = Object.keys(TIERS)

export const tierOf = (tier) => TIERS[tier] ?? TIERS.general

// "ลูกค้าใหม่" = added within the last 7 × 24 hours (counted from created_at, not calendar days).
// Fixed for now; must match NewCustomerDays in backend/internal/customer/service.go (GET /customers?joined=).
export const NEW_CUSTOMER_DAYS = 7

export const isNewCustomer = (customer) =>
  Boolean(customer.created_at) && Date.now() - new Date(customer.created_at).getTime() < NEW_CUSTOMER_DAYS * 24 * 60 * 60 * 1000

export const fullName = (customer) => [customer.fname, customer.lname].filter(Boolean).join(' ')

// "สมหญิง ใจดี (หญิง)" — for lists where staff look people up by nickname
export const nameWithNickname = (customer) =>
  customer.nickname ? `${fullName(customer)} (${customer.nickname})` : fullName(customer)

// Whole years since b_day ("2000-01-31" or "2000-01-31T00:00:00Z"); null when empty or in the future.
// Uses only the date part so the time zone can't shift the birthday by a day.
export const ageOf = (bDay) => {
  if (!bDay) return null
  const born = dayjs(bDay.slice(0, 10))
  if (!born.isValid()) return null
  const today = dayjs().startOf('day')
  // diff() rounds toward zero, so a date a few days ahead would give 0 instead of "not born yet"
  if (born.isAfter(today)) return null
  return today.diff(born, 'year')
}

// "31 ม.ค. 2543 (26 ปี)"-style suffix for screens that show the birthday
export const withAge = (text, bDay) => {
  const age = ageOf(bDay)
  return age == null ? text : `${text} (${age} ปี)`
}

export const initialOf = (customer) => customer.fname?.trim().charAt(0) || '?'

// Show only the last 4 digits on screen; the full number stays in the edit form
export const maskCitizenId = (value) => (value ? `•••••••••${value.slice(-4)}` : '')
