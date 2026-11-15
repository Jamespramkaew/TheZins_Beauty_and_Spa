import { FiCheck } from 'react-icons/fi'
import { Link } from 'react-router-dom'
import { formatDateOnly } from '../../../utils/date'
import { formatBaht, formatNumber } from '../../../utils/format'
import { toNumber } from '../../../utils/number'
import { itemName, lineTotal, methodLabel } from '../orderUtils'

// Shown after POST /orders succeeds; order = the created bill as the backend returned it
function SaleReceipt({ order, customer, onNewSale }) {
  const outstanding = toNumber(order.outstanding)

  return (
    <div className="mx-auto max-w-[560px] rounded-card border border-line bg-white p-6">
      <div className="flex flex-col items-center text-center">
        <span className="flex size-12 items-center justify-center rounded-full bg-success-bg text-success">
          <FiCheck className="size-6" />
        </span>
        <h1 className="mt-3 text-2xl font-bold text-ink">ขายสำเร็จ</h1>
        <p className="mt-1 text-sm text-muted">
          เลขที่บิล <span className="font-semibold text-ink">{order.order_no}</span> · {formatDateOnly(order.order_date)}
        </p>
        <p className="mt-0.5 text-sm text-muted">{[customer.fname, customer.lname].filter(Boolean).join(' ')}</p>
      </div>

      <ul className="mt-6 divide-y divide-line border-y border-line text-sm">
        {order.items.map((item) => (
          <li key={item.id} className="flex justify-between gap-4 py-3">
            <span className="min-w-0 text-ink [overflow-wrap:anywhere]">
              {itemName(item)}
              {item.item_type === 'stock' && ` × ${formatNumber(item.qty)}`}
              {toNumber(item.discount) > 0 && (
                <span className="block text-xs text-muted">ส่วนลด −{formatBaht(toNumber(item.discount))}</span>
              )}
            </span>
            <span className="shrink-0 font-medium">{formatBaht(lineTotal(item))}</span>
          </li>
        ))}
      </ul>

      <dl className="mt-4 space-y-1.5 text-sm">
        <div className="flex justify-between font-semibold text-ink">
          <dt>ยอดสุทธิ</dt>
          <dd>{formatBaht(toNumber(order.total))}</dd>
        </div>
        {order.payments.map((payment) => (
          <div key={payment.id} className="flex justify-between text-muted">
            <dt>ชำระ · {methodLabel(payment.method)}</dt>
            <dd>{formatBaht(toNumber(payment.amount))}</dd>
          </div>
        ))}
        <div className={`flex justify-between ${outstanding > 0 ? 'font-semibold text-danger' : 'text-muted'}`}>
          <dt>ค้างชำระ</dt>
          <dd>{formatBaht(outstanding)}</dd>
        </div>
      </dl>

      <div className="mt-6 grid gap-2.5 sm:grid-cols-3">
        <Link
          to={`/orders/${order.id}`}
          className="flex h-11 items-center justify-center rounded-control border border-line-strong text-sm font-medium text-ink transition hover:bg-surface-2"
        >
          ดูบิล
        </Link>
        <Link
          to={`/customers/${order.customer_id}`}
          className="flex h-11 items-center justify-center rounded-control border border-line-strong text-sm font-medium text-ink transition hover:bg-surface-2"
        >
          กลับหน้าลูกค้า
        </Link>
        <button
          type="button"
          onClick={onNewSale}
          className="h-11 rounded-control bg-primary text-sm font-semibold text-white shadow-sm transition hover:bg-[#4d1621]"
        >
          ขายรายการใหม่
        </button>
      </div>
    </div>
  )
}

export default SaleReceipt
