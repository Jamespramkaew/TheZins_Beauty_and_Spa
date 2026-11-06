import { readApiError } from '../../utils/apiError'

// Backend detail (English) → [{ field, message }] for react-hook-form setError.
// Empty array = no field to point at, so the caller shows a generic banner.
const RULES = [
  [/insufficient stock: available (\d+)/, 'qty', (m) => `คงเหลือแค่ ${m[1]} ชิ้น — ยอดอาจเปลี่ยนไประหว่างที่กรอก`],
  [/note is required/, 'note', () => 'กรุณาระบุเหตุผลของการปรับยอด'],
  [/qty must not be zero/, 'qty', () => 'จำนวนต้องไม่เป็น 0'],
  [/qty .*(positive|greater than zero)/, 'qty', () => 'จำนวนต้องมากกว่า 0'],
  [/employee is not active/, 'employee_id', () => 'พนักงานคนนี้ถูกปิดใช้งานแล้ว — เลือกคนอื่น'],
  [/employee does not belong/, 'employee_id', () => 'พนักงานคนนี้ไม่ได้อยู่สาขาที่เลือก'],
  [/employee_id (not found|is required|must be)/, 'employee_id', () => 'กรุณาเลือกพนักงานที่ทำรายการ'],
  [/stock_id (not found|is required|must be)/, 'stock_id', () => 'ไม่พบสินค้านี้ในระบบแล้ว — เลือกสินค้าอื่น'],
  [/origin and destination branch must be different/, 'to_branch_id', () => 'สาขาปลายทางต้องไม่ใช่สาขาต้นทาง'],
  [/destination branch/, 'to_branch_id', () => 'สาขาปลายทางนี้ใช้งานไม่ได้แล้ว'],
  [/(origin branch|branch is not active|branch_id not found)/, 'branch_id', () => 'สาขานี้ใช้งานไม่ได้แล้ว'],
  [/stock name must not exceed/, 'name', () => 'ชื่อสินค้ายาวได้ไม่เกิน 150 ตัวอักษร'],
  [/stock name (is required|cannot be empty)/, 'name', () => 'กรุณากรอกชื่อสินค้า'],
]

export function mapStockError(error) {
  const { status, message, detail } = readApiError(error)

  if (status === 409 || message === 'Stock name already exists') {
    return [{ field: 'name', message: 'มีสินค้าชื่อนี้อยู่แล้ว' }]
  }
  for (const [pattern, field, toMessage] of RULES) {
    const match = detail.match(pattern)
    if (match) return [{ field, message: toMessage(match) }]
  }
  return []
}

// Message for a whole-page load failure
export function loadErrorText(error) {
  const { status } = readApiError(error)
  if (status === 403) {
    return { title: 'บัญชีนี้ยังไม่ได้ผูกกับสาขา', detail: 'ติดต่อผู้ดูแลระบบให้กำหนดสาขาให้บัญชีนี้ก่อนใช้งานหน้าสต็อก' }
  }
  return { title: 'โหลดข้อมูลไม่สำเร็จ', detail: 'ตรวจสอบการเชื่อมต่ออินเทอร์เน็ต แล้วลองใหม่อีกครั้ง' }
}
