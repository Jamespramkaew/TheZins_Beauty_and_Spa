import { NEW_CUSTOMER_DAYS, isNewCustomer } from '../customerUtils'

// Shown only while the customer is new; disappears by itself after NEW_CUSTOMER_DAYS
function NewCustomerBadge({ customer }) {
  if (!isNewCustomer(customer)) return null

  return (
    <span
      title={`เพิ่มเข้าระบบไม่เกิน ${NEW_CUSTOMER_DAYS} วัน`}
      className="inline-flex items-center gap-1.5 rounded-md bg-success-bg px-2 py-0.5 text-xs font-medium text-success"
    >
      <span aria-hidden="true" className="size-1.5 rounded-full bg-success" />
      ลูกค้าใหม่
    </span>
  )
}

export default NewCustomerBadge
