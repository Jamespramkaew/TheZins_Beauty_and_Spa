import { readApiError } from '../../utils/apiError'
import { formatBaht } from '../../utils/format'

// Backend detail (English) → [{ field, message }] for react-hook-form setError.
// Empty array = no field to point at, so the form shows only the generic banner.
const PAYMENT_RULES = [
  [
    /amount ([\d.]+) exceeds outstanding ([\d.]+)/,
    (m) => ({ field: 'amount', message: `ยอดค้างเหลือ ${formatBaht(Number(m[2]))} — อาจมีคนรับชำระไปก่อนหน้านี้` }),
  ],
  [/amount must be more than 0/, () => ({ field: 'amount', message: 'จำนวนเงินต้องมากกว่า 0 และมีทศนิยมไม่เกิน 2 ตำแหน่ง' })],
  [/method must be/, () => ({ field: 'method', message: 'กรุณาเลือกวิธีชำระ' })],
  [/reference must be at most (\d+)/, (m) => ({ field: 'reference', message: `เลขอ้างอิงยาวได้ไม่เกิน ${m[1]} ตัวอักษร` })],
]

// POST /orders errors → { line, message } (line = index in the bill, null = whole bill)
const CREATE_RULES = [
  [/items\[(\d+)\]: discount must be between 0 and ([\d.]+)/, (m) => `ส่วนลดได้ไม่เกิน ${formatBaht(Number(m[2]))}`],
  [/items\[(\d+)\]: course is not active/, () => 'คอร์สนี้ถูกปิดขายไปแล้ว — นำออกจากบิล'],
  [/items\[(\d+)\]: course not found/, () => 'ไม่พบคอร์สนี้แล้ว — นำออกจากบิล'],
  [/items\[(\d+)\]: stock is not for sale/, () => 'สินค้านี้ไม่ได้เปิดขายแล้ว — นำออกจากบิล'],
  [/items\[(\d+)\]: stock not found/, () => 'ไม่พบสินค้านี้แล้ว — นำออกจากบิล'],
]

export function mapCreateOrderError(error) {
  const { status, detail } = readApiError(error)

  if (status === 403) return { line: null, message: 'บัญชีนี้สร้างบิลไม่ได้ — ต้องใช้บัญชีสาขา' }
  for (const [pattern, toMessage] of CREATE_RULES) {
    const match = detail.match(pattern)
    if (match) return { line: Number(match[1]), message: toMessage(match) }
  }
  const overpaid = detail.match(/total payments ([\d.]+) exceed order total ([\d.]+)/)
  if (overpaid) return { line: null, message: `รับชำระได้ไม่เกินยอดสุทธิ ${formatBaht(Number(overpaid[2]))}` }
  if (/customer not found/.test(detail)) return { line: null, message: 'ไม่พบลูกค้าคนนี้แล้ว — เลือกลูกค้าใหม่' }
  return { line: null, message: 'บันทึกไม่สำเร็จ กรุณาลองใหม่อีกครั้ง' }
}

export function mapPaymentError(error) {
  const { detail } = readApiError(error)

  for (const [pattern, toError] of PAYMENT_RULES) {
    const match = detail.match(pattern)
    if (match) return [toError(match)]
  }
  return []
}
