import { formatNumber } from '../../utils/format'

export const PAGE_SIZE_OPTIONS = [10, 20, 50]
export const DEFAULT_CATEGORIES = ['facial', 'laser', 'massage', 'body']

// price / service_comm come from the API as decimal strings
export const toNumber = (value) => {
  const number = Number(value)
  return Number.isFinite(number) ? number : 0
}

export const pricePerSession = (price, maxSession) =>
  Number(maxSession) > 0 ? Math.round(toNumber(price) / Number(maxSession)) : null

export const formatValidDay = (validDay) => (validDay == null ? '∞ ไม่หมดอายุ' : `${formatNumber(validDay)} วัน`)

export const categoryLabel = (category) => (category ? category.charAt(0).toUpperCase() + category.slice(1) : '')

const CATEGORY_TONES = {
  facial: 'bg-[#F3EDF7] text-[#5B3A7A]',
  laser: 'bg-[#EEF1F8] text-chart',
  massage: 'bg-gold-bg text-gold-ink',
  body: 'bg-[#E7F3F1] text-[#1F5E57]',
}

export const categoryTone = (category) => CATEGORY_TONES[category?.toLowerCase()] ?? 'bg-surface-2 text-muted'

// uses_per_unit > 1 = opened as a container, use_qty is capped at uses_per_unit.
// Unknown stock (e.g. deleted since) is treated as a single item with no cap.
export const isContainer = (stock) => (stock?.uses_per_unit ?? 1) > 1

export const maxUseQty = (stock) => (isContainer(stock) ? stock.uses_per_unit : Infinity)
