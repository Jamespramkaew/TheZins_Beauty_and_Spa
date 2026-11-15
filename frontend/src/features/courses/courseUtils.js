import { formatNumber } from '../../utils/format'

export const PAGE_SIZE_OPTIONS = [10, 20, 50]
export const DEFAULT_CATEGORIES = ['facial', 'massage', 'body', 'spa', 'wax', 'all']

// price / service_comm come from the API as decimal strings
export const toNumber = (value) => {
  const number = Number(value)
  return Number.isFinite(number) ? number : 0
}

export const pricePerSession = (price, maxSession) =>
  Number(maxSession) > 0 ? Math.round(toNumber(price) / Number(maxSession)) : null

export const formatValidDay = (validDay) => (validDay == null ? '∞ ไม่หมดอายุ' : `${formatNumber(validDay)} วัน`)

// Filter-bar value for the "ปิดขาย" button. Not a real category (categories are typed freely,
// so a plain word like "inactive" could collide); it maps to ?active=false instead of ?category=.
export const INACTIVE_FILTER = '__inactive'

// Courses per category among `courses`, for the filter badges ('' = all of them, INACTIVE_FILTER = ปิดขาย).
// Follows the search box like the list does: name contains `search`, ignoring case.
export const countByCategory = (courses, search) => {
  const needle = search.toLowerCase()
  const counts = { '': 0, [INACTIVE_FILTER]: 0 }
  for (const course of courses) {
    if (needle && !course.name.toLowerCase().includes(needle)) continue
    counts[''] += 1
    if (!course.is_active) counts[INACTIVE_FILTER] += 1
    const category = course.category?.trim().toLowerCase()
    if (category) counts[category] = (counts[category] ?? 0) + 1
  }
  return counts
}

export const categoryLabel = (category) => (category ? category.charAt(0).toUpperCase() + category.slice(1) : '')

const CATEGORY_TONES = {
  facial: 'bg-[#F3EDF7] text-[#5B3A7A]',
  massage: 'bg-gold-bg text-gold-ink',
  body: 'bg-[#E7F3F1] text-[#1F5E57]',
  spa: 'bg-[#EEF1F8] text-chart',
  wax: 'bg-[#FBEFEA] text-[#8A3F22]',
  all: 'bg-primary/10 text-primary',
}

export const categoryTone = (category) => CATEGORY_TONES[category?.toLowerCase()] ?? 'bg-surface-2 text-muted'

// uses_per_unit > 1 = opened as a container, use_qty is capped at uses_per_unit.
// Unknown stock (e.g. deleted since) is treated as a single item with no cap.
export const isContainer = (stock) => (stock?.uses_per_unit ?? 1) > 1

export const maxUseQty = (stock) => (isContainer(stock) ? stock.uses_per_unit : Infinity)
