import { useState } from 'react'
import { FiArrowLeft, FiCreditCard, FiRefreshCw } from 'react-icons/fi'
import { Link, useParams } from 'react-router-dom'
import { orderApi } from '../../api/orderApi'
import OrderStatusBadge from '../../components/OrderStatusBadge'
import { useQuery } from '../../hooks/useQuery'
import { readApiError } from '../../utils/apiError'
import { formatDateOnly, formatDateTime } from '../../utils/date'
import { formatBaht, formatNumber } from '../../utils/format'
import { toNumber } from '../../utils/number'
import PaymentModal from './components/PaymentModal'
import { itemName, itemTypeLabel, lineTotal, methodLabel, sourceLabel } from './orderUtils'

// useQuery needs a stable fetcher, so the id goes in params
const fetchOrder = ({ id }) => orderApi.getById(id)

function OrderDetailPage() {
  const { id } = useParams()
  const [paying, setPaying] = useState(false)
  const query = useQuery(fetchOrder, { id })

  if (query.error) {
    const notFound = readApiError(query.error).status === 404
    return (
      <PageShell>
        <div
          role="alert"
          className="flex flex-col items-center justify-center rounded-card border border-dashed border-line-strong bg-surface px-6 py-16 text-center"
        >
          <p className="font-semibold text-ink">{notFound ? 'ไม่พบบิลนี้' : 'โหลดบิลไม่สำเร็จ'}</p>
          <p className="mt-1 text-sm text-muted">
            {notFound ? 'ลิงก์อาจไม่ถูกต้อง หรือเป็นบิลของสาขาอื่น' : 'ตรวจสอบการเชื่อมต่อ แล้วลองใหม่อีกครั้ง'}
          </p>
          {!notFound && (
            <button
              type="button"
              onClick={query.reload}
              className="mt-4 flex h-10 items-center gap-2 rounded-control border border-line-strong bg-white px-4 text-sm font-medium text-ink hover:bg-surface-2"
            >
              <FiRefreshCw className="size-4" />
              ลองใหม่
            </button>
          )}
        </div>
      </PageShell>
    )
  }

  if (!query.loaded) {
    return (
      <PageShell>
        <p className="rounded-card border border-dashed border-line-strong bg-surface px-6 py-16 text-center font-semibold text-ink">
          กำลังโหลดบิล…
        </p>
      </PageShell>
    )
  }

  const order = query.data
  const outstanding = toNumber(order.outstanding)
  const customerName = [order.customer?.fname, order.customer?.lname].filter(Boolean).join(' ')

  return (
    <PageShell>
      <header className="flex flex-wrap items-start justify-between gap-4 rounded-card border border-line bg-white p-5">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <h1 className="text-[26px] font-bold tracking-tight text-ink">{order.order_no}</h1>
            <OrderStatusBadge status={order.status} />
          </div>
          <p className="mt-1 flex flex-wrap gap-x-4 gap-y-1 text-sm text-muted">
            <span>ขายวันที่ {formatDateOnly(order.order_date)}</span>
            {order.branch && <span>{order.branch.name}</span>}
            {order.source && <span>ช่องทาง {sourceLabel(order.source)}</span>}
          </p>
          {order.customer && (
            <p className="mt-2 text-sm">
              ลูกค้า{' '}
              <Link to={`/customers/${order.customer_id}`} className="font-semibold text-chart underline-offset-4 hover:underline">
                {customerName}
              </Link>{' '}
              <span className="text-muted">· {order.customer.hn_no}</span>
            </p>
          )}
        </div>

        {outstanding > 0 && (
          <button
            type="button"
            onClick={() => setPaying(true)}
            className="flex h-11 items-center gap-2 rounded-control bg-primary px-4 text-sm font-semibold text-white shadow-sm transition hover:bg-[#4d1621]"
          >
            <FiCreditCard className="size-4" />
            รับชำระเพิ่ม
          </button>
        )}
      </header>

      <dl className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <Amount label="ยอดรวม" value={order.total} />
        <Amount label="ชำระแล้ว" value={order.paid} tone="text-success" />
        <Amount label="ค้างชำระ" value={order.outstanding} tone={outstanding > 0 ? 'text-danger' : 'text-muted'} />
      </dl>

      <section className="space-y-3">
        <h2 className="text-[15px] font-semibold text-ink">รายการในบิล</h2>
        <div className="overflow-x-auto rounded-card border border-line">
          <table className="w-full min-w-[720px] text-sm">
            <thead className="bg-surface text-xs text-muted">
              <tr className="text-left">
                <th className="px-4 py-3 font-medium">#</th>
                <th className="px-3 py-3 font-medium">รายการ</th>
                <th className="px-3 py-3 text-right font-medium">จำนวน</th>
                <th className="px-3 py-3 text-right font-medium">ราคาต่อหน่วย</th>
                <th className="px-3 py-3 text-right font-medium">ส่วนลด</th>
                <th className="px-4 py-3 text-right font-medium">รวม</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line bg-white">
              {order.items.map((item) => (
                <tr key={item.id}>
                  <td className="px-4 py-3.5 text-muted">{item.line_no}</td>
                  <td className="px-3 py-3.5">
                    <p className="font-medium text-ink [overflow-wrap:anywhere]">{itemName(item)}</p>
                    <p className="mt-0.5 text-xs text-muted">
                      {itemTypeLabel(item.item_type)}
                      {item.total_session != null && ` · ${formatNumber(item.total_session)} ครั้ง`}
                      {item.expires_at && ` · หมดอายุ ${formatDateOnly(item.expires_at)}`}
                    </p>
                  </td>
                  <td className="px-3 py-3.5 text-right whitespace-nowrap">{formatNumber(item.qty)}</td>
                  <td className="px-3 py-3.5 text-right whitespace-nowrap">{formatBaht(toNumber(item.unit_price))}</td>
                  <td className="px-3 py-3.5 text-right whitespace-nowrap text-muted">
                    {toNumber(item.discount) > 0 ? `−${formatBaht(toNumber(item.discount))}` : '–'}
                  </td>
                  <td className="px-4 py-3.5 text-right font-semibold whitespace-nowrap">{formatBaht(lineTotal(item))}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section className="space-y-3">
        <h2 className="text-[15px] font-semibold text-ink">ประวัติการชำระ</h2>
        {order.payments.length === 0 ? (
          <p className="rounded-card border border-dashed border-line-strong bg-surface px-6 py-10 text-center text-sm text-muted">
            ยังไม่มีการชำระ
          </p>
        ) : (
          <div className="overflow-x-auto rounded-card border border-line">
            <table className="w-full min-w-[560px] text-sm">
              <thead className="bg-surface text-xs text-muted">
                <tr className="text-left">
                  <th className="px-4 py-3 font-medium">วันเวลา</th>
                  <th className="px-3 py-3 font-medium">วิธีชำระ</th>
                  <th className="px-3 py-3 font-medium">เลขอ้างอิง</th>
                  <th className="px-4 py-3 text-right font-medium">จำนวนเงิน</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-line bg-white">
                {order.payments.map((payment) => (
                  <tr key={payment.id}>
                    <td className="px-4 py-3.5 whitespace-nowrap">{formatDateTime(payment.paid_at)}</td>
                    <td className="px-3 py-3.5 whitespace-nowrap">{methodLabel(payment.method)}</td>
                    <td className="px-3 py-3.5 text-muted [overflow-wrap:anywhere]">{payment.reference || '–'}</td>
                    <td className="px-4 py-3.5 text-right font-semibold whitespace-nowrap">
                      {formatBaht(toNumber(payment.amount))}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      {paying && (
        <PaymentModal
          order={order}
          onClose={() => setPaying(false)}
          onPaid={() => {
            setPaying(false)
            query.reload()
          }}
        />
      )}
    </PageShell>
  )
}

function PageShell({ children }) {
  return (
    <div className="mx-auto max-w-[1100px] space-y-6">
      <Link
        to="/orders"
        className="inline-flex items-center gap-1.5 text-sm text-muted underline-offset-4 hover:text-ink hover:underline"
      >
        <FiArrowLeft className="size-4" />
        บิลทั้งหมด
      </Link>
      {children}
    </div>
  )
}

function Amount({ label, value, tone = 'text-ink' }) {
  return (
    <div className="rounded-card border border-line bg-white p-5">
      <dt className="text-sm text-muted">{label}</dt>
      <dd className={`mt-2 text-[24px] leading-none font-bold tracking-tight ${tone}`}>{formatBaht(toNumber(value))}</dd>
    </div>
  )
}

export default OrderDetailPage
