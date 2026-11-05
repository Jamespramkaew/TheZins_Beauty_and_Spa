export const PAGE_SIZE_OPTIONS = [10, 20, 50]

// value '' = every status (GET /orders without status)
export const STATUS_FILTERS = [
  { value: '', label: 'ทั้งหมด' },
  { value: 'unpaid', label: 'ยังไม่ชำระ' },
  { value: 'partial', label: 'ชำระบางส่วน' },
  { value: 'paid', label: 'ชำระครบ' },
]
