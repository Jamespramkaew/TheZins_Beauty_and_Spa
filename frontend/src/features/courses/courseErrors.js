import { readApiError } from '../../utils/apiError'

const ITEM_RULES = [
  [/items\[(\d+)\]\.stock_id '.*' not found/, () => 'ไม่พบสินค้านี้ในระบบแล้ว — กด × เพื่อนำออก แล้วเลือกสินค้าใหม่'],
  [
    /items\[(\d+)\]\.use_qty must not exceed (\d+) \(uses per unit of '(.+)'\)/,
    (match) => `ใส่ได้สูงสุด ${match[2]} ครั้ง เพราะ 1 กระปุกของ “${match[3]}” ใช้ได้ ${match[2]} ครั้ง`,
  ],
  [/items\[(\d+)\]\.stock_id '.*' is duplicated/, () => 'เลือกสินค้านี้ซ้ำกับแถวอื่น — นำออกให้เหลือแถวเดียว'],
]

// Returns [{ field, message }] for react-hook-form setError. Empty = show only the generic banner.
export function mapCourseError(error) {
  const { status, message, detail } = readApiError(error)

  for (const [pattern, toMessage] of ITEM_RULES) {
    const match = detail.match(pattern)
    if (match) return [{ field: `items.${match[1]}.stock_id`, message: toMessage(match) }]
  }
  if (status === 409 || message === 'Course name already exists') {
    return [{ field: 'name', message: 'มีคอร์สชื่อนี้อยู่แล้ว — ลองเปลี่ยนชื่อ' }]
  }
  return []
}
