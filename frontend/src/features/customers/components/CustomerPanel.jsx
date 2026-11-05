import { FiArrowRight, FiEdit2, FiStar, FiX } from 'react-icons/fi'
import { Link } from 'react-router-dom'
import { formatDate, formatDateOnly } from '../../../utils/date'
import { fullName, initialOf, maskCitizenId, tierOf } from '../customerUtils'
import NewCustomerBadge from './NewCustomerBadge'
import TierBadge from './TierBadge'

function CustomerPanel({ customer, onClose, onEdit, onChangeTier }) {
  const name = fullName(customer)

  return (
    <aside aria-label="รายละเอียดลูกค้า" className="rounded-card border border-line bg-white xl:sticky xl:top-6">
      <div className="relative flex flex-col items-center border-b border-line px-5 pt-6 pb-5 text-center">
        <button
          type="button"
          onClick={onClose}
          aria-label="ปิดรายละเอียด"
          className="absolute top-3 right-3 flex size-9 items-center justify-center rounded-control text-muted transition hover:bg-surface-2 hover:text-ink"
        >
          <FiX className="size-4" />
        </button>
        <span
          aria-hidden="true"
          className={`flex size-16 items-center justify-center rounded-full text-2xl font-semibold ${
            tierOf(customer.tier).avatarClassName
          }`}
        >
          {initialOf(customer)}
        </span>
        <p className="mt-3 text-lg font-bold text-ink [overflow-wrap:anywhere]">{name}</p>
        <div className="mt-1.5 flex flex-wrap justify-center gap-1.5">
          <TierBadge tier={customer.tier} />
          <NewCustomerBadge customer={customer} />
        </div>
      </div>

      <dl className="divide-y divide-line px-5 text-sm">
        <Row label="เลข HN" value={customer.hn_no} />
        <Row label="เบอร์โทร" value={customer.tel} />
        <Row label="เลขบัตรประชาชน" value={maskCitizenId(customer.citizen_id)} />
        <Row label="วันเกิด" value={customer.b_day ? formatDateOnly(customer.b_day) : ''} />
        <Row label="ที่อยู่" value={customer.address} />
        <Row label="วันที่เพิ่ม" value={formatDate(customer.created_at)} />
      </dl>

      <div className="grid grid-cols-2 gap-2.5 border-t border-line p-5">
        <button
          type="button"
          onClick={() => onEdit(customer)}
          className="flex h-10 items-center justify-center gap-2 rounded-control border border-line-strong bg-white text-sm font-medium text-ink transition hover:bg-surface-2"
        >
          <FiEdit2 className="size-4" />
          แก้ไขข้อมูล
        </button>
        <button
          type="button"
          onClick={() => onChangeTier(customer)}
          className="flex h-10 items-center justify-center gap-2 rounded-control border border-line-strong bg-white text-sm font-medium text-ink transition hover:bg-surface-2"
        >
          <FiStar className="size-4" />
          {customer.tier === 'vip' ? 'เปลี่ยนเป็นทั่วไป' : 'เปลี่ยนเป็น VIP'}
        </button>
        <Link
          to={`/customers/${customer.id}`}
          className="col-span-2 flex h-10 items-center justify-center gap-2 rounded-control bg-primary text-sm font-semibold text-white shadow-sm transition hover:bg-[#4d1621]"
        >
          ดูประวัติการซื้อและรายละเอียด
          <FiArrowRight className="size-4" />
        </Link>
      </div>
    </aside>
  )
}

function Row({ label, value }) {
  return (
    <div className="flex justify-between gap-4 py-3">
      <dt className="shrink-0 text-muted">{label}</dt>
      <dd className="text-right text-ink [overflow-wrap:anywhere]">{value || '–'}</dd>
    </div>
  )
}

export default CustomerPanel
