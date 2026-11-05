import { useState } from 'react'
import { FiArrowLeft, FiInfo } from 'react-icons/fi'
import { Link, useSearchParams } from 'react-router-dom'
import { courseApi } from '../../api/courseApi'
import { customerApi } from '../../api/customerApi'
import { orderApi } from '../../api/orderApi'
import { stockApi } from '../../api/stockApi'
import { useAuth } from '../../hooks/useAuth'
import { useQuery } from '../../hooks/useQuery'
import { useToast } from '../../hooks/useToast'
import { toISODate } from '../../utils/date'
import { toNumber } from '../../utils/number'
import BillPanel from './components/BillPanel'
import CatalogPanel from './components/CatalogPanel'
import CustomerPicker from './components/CustomerPicker'
import SaleReceipt from './components/SaleReceipt'
import { mapCreateOrderError } from './orderErrors'
import { MONEY_PATTERN } from './orderUtils'

// useQuery needs stable fetchers — keep them at module level
const fetchCustomer = ({ id }) => customerApi.getById(id)

// Every active course (GET /courses caps page_size at 100, so walk the pages)
async function fetchActiveCourses() {
  const all = []
  for (let page = 1; ; page++) {
    const response = await courseApi.getAll({ page, page_size: 100 })
    all.push(...response.data)
    if (page >= (response.meta?.total_pages ?? 1)) return { data: all.filter((course) => course.is_active) }
  }
}

// Products that can be sold: active and is_sellable (the backend rejects the rest)
const fetchSellableStocks = () =>
  stockApi.getAll().then((response) => ({ data: (response.data ?? []).filter((stock) => stock.is_active && stock.is_sellable) }))

const EMPTY_PAYMENT = { method: 'cash', amount: '', touched: false }

// Prices come from the catalog; the backend uses the same catalog price (unit_price is never sent).
// /sell?customer=<id> opens the bill for that customer; without it the page asks who it's for first.
function SellPage() {
  const { user } = useAuth()
  const toast = useToast()
  const [searchParams, setSearchParams] = useSearchParams()
  const customerId = searchParams.get('customer') ?? ''
  const today = toISODate(new Date())

  const [lines, setLines] = useState([])
  const [nextKey, setNextKey] = useState(1)
  const [payment, setPayment] = useState(EMPTY_PAYMENT)
  const [orderDate, setOrderDate] = useState(today)
  const [source, setSource] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState(null) // { line, message }
  const [created, setCreated] = useState(null)

  const isBranch = user.role === 'branch'
  const customerQuery = useQuery(fetchCustomer, { id: customerId }, { enabled: isBranch && Boolean(customerId) })
  const courses = useQuery(fetchActiveCourses, {}, { enabled: isBranch })
  const stocks = useQuery(fetchSellableStocks, {}, { enabled: isBranch })

  // Admin can't create orders (POST /orders is branch only) — say so instead of a form that would fail
  if (!isBranch) {
    return (
      <PageShell>
        <p className="flex items-start gap-3 rounded-card border border-line bg-surface px-5 py-4 text-sm text-muted">
          <FiInfo className="mt-0.5 size-4 shrink-0" />
          การขายต้องทำจากบัญชีสาขา เพราะบิลต้องผูกกับสาขาที่ขาย — บัญชี admin ดูบิลทุกสาขาได้ที่หน้า Orders
        </p>
      </PageShell>
    )
  }

  const totals = lines.reduce(
    (sum, line) => ({
      gross: sum.gross + line.qty * line.price,
      discount: sum.discount + line.discountValue,
      net: sum.net + line.total,
    }),
    { gross: 0, discount: 0, net: 0 },
  )

  // Until the cashier types an amount, "pay now" follows the net total
  const amountText = payment.touched ? payment.amount : totals.net > 0 ? String(totals.net) : ''
  const amountValue = MONEY_PATTERN.test(amountText.trim()) ? Number(amountText) : NaN
  const amountValid = amountText.trim() === '' || (amountValue > 0 && amountValue <= totals.net)
  const paymentView = {
    method: payment.method,
    amount: amountText,
    valid: amountValid,
    remaining: Math.max(0, totals.net - (amountValid && amountText.trim() !== '' ? amountValue : 0)),
  }

  const withTotals = (line) => {
    const discountText = line.discount.trim()
    const discountValue = MONEY_PATTERN.test(discountText) ? Math.min(Number(discountText), line.qty * line.price) : 0
    return { ...line, discountValue, total: line.qty * line.price - discountValue }
  }

  const updateLine = (key, patch) => {
    setError(null)
    setLines((current) => current.map((line) => (line.key === key ? withTotals({ ...line, ...patch }) : line)))
  }

  const addLine = (line) => {
    setError(null)
    setLines((current) => [...current, withTotals({ ...line, key: nextKey, discount: '' })])
    setNextKey((key) => key + 1)
  }

  // Courses are sold one at a time (backend: course qty must be 1), so the card toggles
  const toggleCourse = (course) => {
    const existing = lines.find((line) => line.type === 'course' && line.id === course.id)
    if (existing) {
      setLines((current) => current.filter((line) => line.key !== existing.key))
      return
    }
    addLine({ type: 'course', id: course.id, name: course.name, price: toNumber(course.price), qty: 1, sessions: course.max_session })
  }

  const addStock = (stock) => {
    const existing = lines.find((line) => line.type === 'stock' && line.id === stock.id)
    if (existing) updateLine(existing.key, { qty: existing.qty + 1 })
    else addLine({ type: 'stock', id: stock.id, name: stock.name, price: toNumber(stock.price), qty: 1 })
  }

  const linesValid = lines.every((line) => {
    const text = line.discount.trim()
    return text === '' || (MONEY_PATTERN.test(text) && Number(text) <= line.qty * line.price)
  })

  const submit = async () => {
    if (!linesValid || !amountValid) {
      setError({ line: null, message: 'แก้ไขจุดที่มีข้อความสีแดงก่อน แล้วกดยืนยันอีกครั้ง' })
      return
    }
    setSubmitting(true)
    setError(null)
    try {
      const response = await orderApi.create({
        customer_id: customerId,
        order_date: orderDate,
        source: source || undefined,
        items: lines.map((line) => ({
          item_type: line.type,
          ...(line.type === 'course' ? { course_id: line.id } : { stock_id: line.id, qty: line.qty }),
          discount: line.discount.trim() || '0',
        })),
        payments: amountText.trim() === '' ? [] : [{ method: payment.method, amount: amountText.trim(), reference: '' }],
      })
      toast.success('ขายสำเร็จ', `บิล ${response.data.order_no}`)
      setCreated(response.data)
    } catch (requestError) {
      // Nothing was saved — keep the bill as it is so the cashier can fix the line
      setError(mapCreateOrderError(requestError))
    } finally {
      setSubmitting(false)
    }
  }

  const resetSale = () => {
    setLines([])
    setPayment(EMPTY_PAYMENT)
    setOrderDate(today)
    setSource('')
    setError(null)
    setCreated(null)
  }

  if (created && customerQuery.data) {
    return <SaleReceipt order={created} customer={customerQuery.data} onNewSale={resetSale} />
  }

  if (!customerId) {
    return (
      <PageShell>
        <div className="max-w-xl space-y-4 rounded-card border border-line bg-white p-5">
          <div>
            <h1 className="text-xl font-bold text-ink">ขาย — เลือกลูกค้า</h1>
            <p className="mt-1 text-sm text-muted">เลือกลูกค้าก่อน แล้วค่อยเลือกคอร์สหรือสินค้า</p>
          </div>
          <CustomerPicker onPick={(customer) => setSearchParams({ customer: customer.id })} />
        </div>
      </PageShell>
    )
  }

  const loadFailed = customerQuery.error || courses.error || stocks.error
  if (loadFailed || !customerQuery.loaded || !courses.loaded || !stocks.loaded) {
    return (
      <PageShell customerId={customerId}>
        <p
          role={loadFailed ? 'alert' : undefined}
          className="rounded-card border border-dashed border-line-strong bg-surface px-6 py-16 text-center text-sm text-muted"
        >
          {loadFailed ? 'โหลดข้อมูลไม่สำเร็จ — ลิงก์ลูกค้าอาจไม่ถูกต้อง หรือการเชื่อมต่อมีปัญหา' : 'กำลังโหลด…'}
        </p>
      </PageShell>
    )
  }

  return (
    <PageShell customerId={customerId}>
      <div>
        <h1 className="text-[28px] font-bold tracking-tight text-ink">ขาย</h1>
        <p className="mt-1 text-sm text-muted">เลือกคอร์สหรือสินค้า แล้วเพิ่มลงบิลทางขวา · ราคาตามที่ตั้งไว้ในระบบ</p>
      </div>

      <div className="grid items-start gap-6 xl:grid-cols-[minmax(0,1fr)_420px]">
        <CatalogPanel
          courses={courses.data}
          stocks={stocks.data}
          lines={lines}
          onToggleCourse={toggleCourse}
          onAddStock={addStock}
        />
        <BillPanel
          customer={customerQuery.data}
          onChangeCustomer={() => {
            resetSale()
            setSearchParams({})
          }}
          lines={lines}
          lineErrors={error?.line != null ? { [error.line]: error.message } : {}}
          onRemove={(key) => {
            setError(null)
            setLines((current) => current.filter((line) => line.key !== key))
          }}
          onQty={(key, qty) => updateLine(key, { qty: Math.max(1, qty) })}
          onDiscount={(key, discount) => updateLine(key, { discount })}
          totals={totals}
          payment={paymentView}
          onPayment={(patch) => {
            setError(null)
            setPayment((current) => ({ ...current, ...patch }))
          }}
          orderDate={orderDate}
          onOrderDate={setOrderDate}
          today={today}
          source={source}
          onSource={setSource}
          error={error?.line == null ? error?.message : null}
          submitting={submitting}
          onSubmit={submit}
        />
      </div>
    </PageShell>
  )
}

function PageShell({ customerId, children }) {
  return (
    <div className="mx-auto max-w-[1400px] space-y-6">
      <Link
        to={customerId ? `/customers/${customerId}` : '/orders'}
        className="inline-flex items-center gap-1.5 text-sm text-muted underline-offset-4 hover:text-ink hover:underline"
      >
        <FiArrowLeft className="size-4" />
        {customerId ? 'กลับหน้าลูกค้า' : 'บิลทั้งหมด'}
      </Link>
      {children}
    </div>
  )
}

export default SellPage
