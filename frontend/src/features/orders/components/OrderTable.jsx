import { Link } from 'react-router-dom'
import OrderStatusBadge from '../../../components/OrderStatusBadge'
import { formatDateOnly } from '../../../utils/date'
import { formatBaht } from '../../../utils/format'
import { toNumber } from '../../../utils/number'

// orders = one page of GET /orders (rows already carry branch_code, customer name/HN and totals)
function OrderTable({ orders }) {
  return (
    <div className="overflow-x-auto rounded-card border border-line">
      <table className="w-full min-w-[960px] text-sm">
        <thead className="bg-surface text-xs text-muted">
          <tr className="text-left">
            <th className="px-4 py-3 font-medium">เลขบิล</th>
            <th className="px-3 py-3 font-medium">วันที่ขาย</th>
            <th className="px-3 py-3 font-medium">ลูกค้า</th>
            <th className="px-3 py-3 font-medium">สาขา</th>
            <th className="px-3 py-3 text-right font-medium">ยอดรวม</th>
            <th className="px-3 py-3 text-right font-medium">ชำระแล้ว</th>
            <th className="px-3 py-3 text-right font-medium">ค้างชำระ</th>
            <th className="px-4 py-3 font-medium">สถานะ</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-line bg-white">
          {orders.map((order) => {
            const outstanding = toNumber(order.outstanding)
            return (
              <tr key={order.id}>
                <td className="px-4 py-3.5 whitespace-nowrap">
                  <Link
                    to={`/orders/${order.id}`}
                    className="font-semibold text-chart underline-offset-4 hover:underline"
                  >
                    {order.order_no}
                  </Link>
                </td>
                <td className="px-3 py-3.5 whitespace-nowrap">{formatDateOnly(order.order_date)}</td>
                <td className="px-3 py-3.5">
                  <Link
                    to={`/customers/${order.customer_id}`}
                    className="font-medium text-chart [overflow-wrap:anywhere] underline-offset-4 hover:underline"
                  >
                    {order.customer_name}
                  </Link>
                  <p className="mt-0.5 text-xs text-muted">{order.customer_hn_no}</p>
                </td>
                <td className="px-3 py-3.5 whitespace-nowrap text-muted">{order.branch_code}</td>
                <td className="px-3 py-3.5 text-right font-semibold whitespace-nowrap">{formatBaht(toNumber(order.total))}</td>
                <td className="px-3 py-3.5 text-right whitespace-nowrap">{formatBaht(toNumber(order.paid))}</td>
                <td
                  className={`px-3 py-3.5 text-right whitespace-nowrap ${outstanding > 0 ? 'font-semibold text-danger' : 'text-muted'}`}
                >
                  {formatBaht(outstanding)}
                </td>
                <td className="px-4 py-3.5">
                  <OrderStatusBadge status={order.status} />
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}

export default OrderTable
