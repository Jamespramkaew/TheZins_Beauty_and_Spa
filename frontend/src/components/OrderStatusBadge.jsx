// status comes from the backend (paid / partial / unpaid), computed from total vs paid
const ORDER_STATUSES = {
  paid: { label: 'ชำระครบ', className: 'bg-success-bg text-success' },
  partial: { label: 'ชำระบางส่วน', className: 'bg-warn-bg text-warn' },
  unpaid: { label: 'ยังไม่ชำระ', className: 'bg-danger-bg text-danger' },
}

function OrderStatusBadge({ status }) {
  const { label, className } = ORDER_STATUSES[status] ?? ORDER_STATUSES.unpaid

  return (
    <span className={`inline-flex rounded-md px-2 py-0.5 text-xs font-medium whitespace-nowrap ${className}`}>
      {label}
    </span>
  )
}

export default OrderStatusBadge
