import { readApiError } from '../../utils/apiError'

const FIELD_LABELS = {
  hn_no: 'เลข HN',
  fname: 'ชื่อ',
  lname: 'นามสกุล',
  tel: 'เบอร์โทร',
}

// Backend detail (English) → [{ field, message }] for react-hook-form setError.
// Empty array = no field to point at, so the form shows only the generic banner.
const RULES = [
  [/^hn_no already exists/, () => ({ field: 'hn_no', message: 'เลข HN นี้มีลูกค้าใช้อยู่แล้ว — ลองเลขอื่น' })],
  [/^(\w+) is required/, (m) => ({ field: m[1], message: `กรุณากรอก${FIELD_LABELS[m[1]] ?? m[1]}` })],
  [
    /^(\w+) must be at most (\d+) characters/,
    (m) => ({ field: m[1], message: `${FIELD_LABELS[m[1]] ?? m[1]}ยาวได้ไม่เกิน ${m[2]} ตัวอักษร` }),
  ],
  [/^citizen_id must be 13 digits/, () => ({ field: 'citizen_id', message: 'เลขบัตรประชาชนต้องเป็นตัวเลข 13 หลัก' })],
  [/^b_day must be YYYY-MM-DD/, () => ({ field: 'b_day', message: 'วันเกิดไม่ถูกต้อง' })],
]

export function mapCustomerError(error) {
  const { detail } = readApiError(error)

  for (const [pattern, toError] of RULES) {
    const match = detail.match(pattern)
    if (match) return [toError(match)]
  }
  return []
}
