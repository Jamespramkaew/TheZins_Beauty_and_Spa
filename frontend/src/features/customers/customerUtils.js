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

export const TIERS = {
  vip: { label: 'VIP', className: 'bg-gold-bg text-gold-ink', avatarClassName: 'bg-gold-bg text-gold-ink' },
  general: { label: 'ทั่วไป', className: 'bg-surface-2 text-muted', avatarClassName: 'bg-surface-2 text-muted' },
}

export const tierOf = (tier) => TIERS[tier] ?? TIERS.general

export const fullName = (customer) => [customer.fname, customer.lname].filter(Boolean).join(' ')

export const initialOf = (customer) => customer.fname?.trim().charAt(0) || '?'

export const formatDate = (value) => (value ? dayjs(value).locale('th').format('D MMM BBBB') : '–')

// Show only the last 4 digits on screen; the full number stays in the edit form
export const maskCitizenId = (value) => (value ? `•••••••••${value.slice(-4)}` : '')
