import Pagination from '../../../components/ui/Pagination'
import { formatBaht } from '../../../utils/format'
import { ORDER_STATUSES, PAGE_SIZE_OPTIONS, formatDateOnly, toNumber } from '../customerUtils'

// orders = one page of GET /orders?customer_id= (the page owns the query so the tab can show its count)
function OrderHistory({ orders, meta, isAdmin, onPageChange, onPageSizeChange }) {
  if (orders.length === 0) {
    return (
      <p className="rounded-card border border-dashed border-line-strong bg-surface px-6 py-12 text-center text-sm text-muted">
        {isAdmin ? 'ลูกค้าคนนี้ยังไม่มีบิล' : 'ลูกค้าคนนี้ยังไม่มีบิลที่สาขานี้'}
      </p>
    )
  }

  return (
    <div className="space-y-4">
      <div className="overflow-x-auto rounded-card border border-line">
        <table className="w-full min-w-[760px] text-sm">
          <thead className="bg-surface text-xs text-muted">
            <tr className="text-left">
              <th className="px-4 py-3 font-medium">เลขบิล</th>
              <th className="px-3 py-3 font-medium">วันที่ขาย</th>
              <th className="px-3 py-3 font-medium">สาขา</th>
              <th className="px-3 py-3 text-right font-medium">ยอดรวม</th>
              <th className="px-3 py-3 text-right font-medium">ชำระแล้ว</th>
              <th className="px-3 py-3 text-right font-medium">ค้างชำระ</th>
              <th className="px-4 py-3 font-medium">สถานะ</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-line bg-white">
            {orders.map((order) => {
              const status = ORDER_STATUSES[order.status] ?? ORDER_STATUSES.unpaid
              const outstanding = toNumber(order.outstanding)
              return (
                <tr key={order.id}>
                  <td className="px-4 py-3.5 font-semibold whitespace-nowrap text-ink">{order.order_no}</td>
                  <td className="px-3 py-3.5 whitespace-nowrap">{formatDateOnly(order.order_date)}</td>
                  <td className="px-3 py-3.5 whitespace-nowrap text-muted">{order.branch_code}</td>
                  <td className="px-3 py-3.5 text-right font-semibold whitespace-nowrap">{formatBaht(toNumber(order.total))}</td>
                  <td className="px-3 py-3.5 text-right whitespace-nowrap">{formatBaht(toNumber(order.paid))}</td>
                  <td className={`px-3 py-3.5 text-right whitespace-nowrap ${outstanding > 0 ? 'font-semibold text-danger' : 'text-muted'}`}>
                    {formatBaht(outstanding)}
                  </td>
                  <td className="px-4 py-3.5">
                    <span className={`inline-flex rounded-md px-2 py-0.5 text-xs font-medium whitespace-nowrap ${status.className}`}>
                      {status.label}
                    </span>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>

      <Pagination
        meta={meta}
        unit="บิล"
        pageSizeOptions={PAGE_SIZE_OPTIONS}
        onPageChange={onPageChange}
        onPageSizeChange={onPageSizeChange}
      />
    </div>
  )
}

export default OrderHistory
