import { FiChevronRight, FiEdit2 } from 'react-icons/fi'
import { formatDate, fullName, initialOf, tierOf } from '../customerUtils'
import NewCustomerBadge from './NewCustomerBadge'
import TierBadge from './TierBadge'

// selectedId = the customer shown in the side panel (highlighted row)
function CustomerTable({ customers, selectedId, onSelect, onEdit }) {
  return (
    <div className="overflow-x-auto rounded-card border border-line">
      <table className="w-full min-w-[720px] text-sm">
        <thead className="bg-surface text-xs text-muted">
          <tr className="text-left">
            <th className="px-4 py-3 font-medium">ชื่อ</th>
            <th className="px-3 py-3 font-medium">HN</th>
            <th className="px-3 py-3 font-medium">เบอร์โทร</th>
            <th className="px-3 py-3 font-medium">วันที่เพิ่ม</th>
            <th className="px-4 py-3 text-right font-medium">การทำงาน</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-line bg-white">
          {customers.map((customer) => (
            <tr
              key={customer.id}
              className={customer.id === selectedId ? 'bg-gold-bg/40 shadow-[inset_3px_0_0_var(--color-gold)]' : ''}
            >
              <td className="px-4 py-3.5">
                <div className="flex items-center gap-3">
                  <span
                    aria-hidden="true"
                    className={`flex size-9 shrink-0 items-center justify-center rounded-full text-sm font-semibold ${
                      tierOf(customer.tier).avatarClassName
                    }`}
                  >
                    {initialOf(customer)}
                  </span>
                  <div className="min-w-0">
                    <button
                      type="button"
                      onClick={() => onSelect(customer)}
                      aria-expanded={customer.id === selectedId}
                      className="rounded-sm text-left font-semibold text-chart [overflow-wrap:anywhere] underline-offset-4 hover:underline focus-visible:ring-3 focus-visible:ring-chart/20 focus-visible:outline-none"
                    >
                      {fullName(customer)}
                    </button>
                    <div className="mt-1 flex flex-wrap gap-1.5">
                      <TierBadge tier={customer.tier} />
                      <NewCustomerBadge customer={customer} />
                    </div>
                  </div>
                </div>
              </td>
              <td className="px-3 py-3.5 whitespace-nowrap text-muted">{customer.hn_no}</td>
              <td className="px-3 py-3.5 whitespace-nowrap">{customer.tel || '–'}</td>
              <td className="px-3 py-3.5 whitespace-nowrap text-muted">{formatDate(customer.created_at)}</td>
              <td className="px-4 py-3.5">
                <div className="flex justify-end gap-2">
                  <button
                    type="button"
                    onClick={() => onEdit(customer)}
                    aria-label={`แก้ไขข้อมูล ${fullName(customer)}`}
                    className="flex size-9 items-center justify-center rounded-control border border-line-strong bg-white text-muted transition hover:bg-surface-2 hover:text-ink"
                  >
                    <FiEdit2 className="size-4" />
                  </button>
                  <button
                    type="button"
                    onClick={() => onSelect(customer)}
                    aria-label={`ดูรายละเอียด ${fullName(customer)}`}
                    className="flex size-9 items-center justify-center rounded-control border border-line-strong bg-white text-muted transition hover:bg-surface-2 hover:text-ink"
                  >
                    <FiChevronRight className="size-4" />
                  </button>
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export default CustomerTable
