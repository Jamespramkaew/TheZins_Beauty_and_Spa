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

export function mapPaymentError(error) {
  const { detail } = readApiError(error)

  for (const [pattern, toError] of PAYMENT_RULES) {
    const match = detail.match(pattern)
    if (match) return [toError(match)]
  }
  return []
}
