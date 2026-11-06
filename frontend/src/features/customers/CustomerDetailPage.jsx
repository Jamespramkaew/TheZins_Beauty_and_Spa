import { useState } from 'react'
import { FiArrowLeft, FiEdit2, FiPhone, FiRefreshCw, FiStar } from 'react-icons/fi'
import { Link, useParams } from 'react-router-dom'
import { customerApi } from '../../api/customerApi'
import { orderApi } from '../../api/orderApi'
import { useAuth } from '../../hooks/useAuth'
import { useQuery } from '../../hooks/useQuery'
import { readApiError } from '../../utils/apiError'
import { formatNumber } from '../../utils/format'
import ChangeTierDialog from './components/ChangeTierDialog'
import CustomerFormModal from './components/CustomerFormModal'
import NewCustomerBadge from './components/NewCustomerBadge'
import OrderHistory from './components/OrderHistory'
import TierBadge from './components/TierBadge'
import TierHistory from './components/TierHistory'
import { PAGE_SIZE_OPTIONS, formatDate, formatDateOnly, fullName, initialOf, maskCitizenId, tierOf } from './customerUtils'

// useQuery needs stable fetchers, so the id goes in params
const fetchCustomer = ({ id }) => customerApi.getById(id)
const fetchTierChanges = ({ id }) => customerApi.getTierChanges(id)

function CustomerDetailPage() {
  const { id } = useParams()
  const { user } = useAuth()
  const isAdmin = user.role === 'admin'
  const [tab, setTab] = useState('orders')
  const [orderPage, setOrderPage] = useState(1)
  const [orderPageSize, setOrderPageSize] = useState(PAGE_SIZE_OPTIONS[0])
  const [editing, setEditing] = useState(false)
  const [changingTier, setChangingTier] = useState(false)
  // Bumped after a tier change so the history tab refetches too
  const [tierRefresh, setTierRefresh] = useState(0)

  const customerQuery = useQuery(fetchCustomer, { id })
  const orders = useQuery(orderApi.getAll, { customer_id: id, page: orderPage, page_size: orderPageSize })
  const tierChanges = useQuery(fetchTierChanges, { id }, { refreshKey: tierRefresh })

  if (customerQuery.error) {
    const notFound = readApiError(customerQuery.error).status === 404
    return (
      <PageShell>
        <div
          role="alert"
          className="flex flex-col items-center justify-center rounded-card border border-dashed border-line-strong bg-surface px-6 py-16 text-center"
        >
          <p className="font-semibold text-ink">{notFound ? 'ไม่พบลูกค้าคนนี้' : 'โหลดข้อมูลลูกค้าไม่สำเร็จ'}</p>
          <p className="mt-1 text-sm text-muted">
            {notFound ? 'ลิงก์อาจไม่ถูกต้อง กลับไปเลือกจากรายชื่อลูกค้า' : 'ตรวจสอบการเชื่อมต่อ แล้วลองใหม่อีกครั้ง'}
          </p>
          {!notFound && (
            <button
              type="button"
              onClick={customerQuery.reload}
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

  if (!customerQuery.loaded) {
    return (
      <PageShell>
        <p className="rounded-card border border-dashed border-line-strong bg-surface px-6 py-16 text-center font-semibold text-ink">
          กำลังโหลดข้อมูลลูกค้า…
        </p>
      </PageShell>
    )
  }

  const customer = customerQuery.data
  const tabs = [
    { value: 'orders', label: 'ประวัติการซื้อ', count: orders.meta?.total },
    { value: 'tiers', label: 'ประวัติระดับ', count: tierChanges.data?.length },
  ]

  return (
    <PageShell>
      <header className="flex flex-wrap items-start justify-between gap-4 rounded-card border border-line bg-white p-5">
        <div className="flex min-w-0 items-center gap-4">
          <span
            aria-hidden="true"
            className={`flex size-16 shrink-0 items-center justify-center rounded-full text-2xl font-semibold ${
              tierOf(customer.tier).avatarClassName
            }`}
          >
            {initialOf(customer)}
          </span>
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2">
              <h1 className="text-[26px] font-bold tracking-tight text-ink [overflow-wrap:anywhere]">{fullName(customer)}</h1>
              <TierBadge tier={customer.tier} />
              <NewCustomerBadge customer={customer} />
            </div>
            <p className="mt-1 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-muted">
              <span>{customer.hn_no}</span>
              {customer.tel && (
                <span className="flex items-center gap-1.5">
                  <FiPhone className="size-3.5" aria-hidden="true" />
                  {customer.tel}
                </span>
              )}
            </p>
          </div>
        </div>

        <div className="flex flex-wrap gap-2.5">
          <button
            type="button"
            onClick={() => setEditing(true)}
            className="flex h-10 items-center gap-2 rounded-control border border-line-strong bg-white px-4 text-sm font-medium text-ink transition hover:bg-surface-2"
          >
            <FiEdit2 className="size-4" />
            แก้ไขข้อมูล
          </button>
          <button
            type="button"
            onClick={() => setChangingTier(true)}
            className="flex h-10 items-center gap-2 rounded-control border border-line-strong bg-white px-4 text-sm font-medium text-ink transition hover:bg-surface-2"
          >
            <FiStar className="size-4" />
            {customer.tier === 'vip' ? 'เปลี่ยนเป็นทั่วไป' : 'เปลี่ยนเป็น VIP'}
          </button>
        </div>
      </header>

      <dl className="grid gap-x-6 gap-y-3 rounded-card border border-line bg-white p-5 text-sm sm:grid-cols-2 xl:grid-cols-4">
        <Info label="เลขบัตรประชาชน" value={maskCitizenId(customer.citizen_id)} />
        <Info label="วันเกิด" value={customer.b_day ? formatDateOnly(customer.b_day) : ''} />
        <Info label="วันที่เพิ่ม" value={formatDate(customer.created_at)} />
        <Info label="ที่อยู่" value={customer.address} />
      </dl>

      <section className="space-y-4">
        <div role="tablist" aria-label="ประวัติลูกค้า" className="flex gap-1 border-b border-line">
          {tabs.map((item) => (
            <button
              key={item.value}
              type="button"
              role="tab"
              aria-selected={tab === item.value}
              onClick={() => setTab(item.value)}
              className={`-mb-px flex h-11 items-center gap-2 border-b-2 px-4 text-sm transition ${
                tab === item.value ? 'border-primary font-semibold text-ink' : 'border-transparent text-muted hover:text-ink'
              }`}
            >
              {item.label}
              {item.count != null && (
                <span className="rounded-full bg-surface-2 px-2 py-0.5 text-xs text-muted">{formatNumber(item.count)}</span>
              )}
            </button>
          ))}
        </div>

        <div role="tabpanel">
          {tab === 'orders' ? (
            <TabBody query={orders} label="ประวัติการซื้อ">
              {!isAdmin && <p className="mb-3 text-xs text-muted">แสดงเฉพาะบิลของสาขานี้</p>}
              <OrderHistory
                orders={orders.data ?? []}
                meta={orders.meta}
                isAdmin={isAdmin}
                onPageChange={setOrderPage}
                onPageSizeChange={(size) => {
                  setOrderPageSize(size)
                  setOrderPage(1)
                }}
              />
            </TabBody>
          ) : (
            <TabBody query={tierChanges} label="ประวัติระดับ">
              <TierHistory rows={tierChanges.data ?? []} />
            </TabBody>
          )}
        </div>
      </section>

      {editing && (
        <CustomerFormModal
          customer={customer}
          onClose={() => setEditing(false)}
          onSaved={() => {
            setEditing(false)
            customerQuery.reload()
          }}
        />
      )}

      {changingTier && (
        <ChangeTierDialog
          customer={customer}
          onClose={() => setChangingTier(false)}
          onChanged={() => {
            setChangingTier(false)
            customerQuery.reload()
            setTierRefresh((value) => value + 1)
          }}
        />
      )}
    </PageShell>
  )
}

function PageShell({ children }) {
  return (
    <div className="mx-auto max-w-[1400px] space-y-6">
      <Link
        to="/customers"
        className="inline-flex items-center gap-1.5 text-sm text-muted underline-offset-4 hover:text-ink hover:underline"
      >
        <FiArrowLeft className="size-4" />
        ลูกค้าทั้งหมด
      </Link>
      {children}
    </div>
  )
}

function Info({ label, value }) {
  return (
    <div className="min-w-0">
      <dt className="text-xs text-muted">{label}</dt>
      <dd className="mt-1 text-ink [overflow-wrap:anywhere]">{value || '–'}</dd>
    </div>
  )
}

// Loading / error wrapper shared by both tabs
function TabBody({ query, label, children }) {
  if (query.error) {
    return (
      <div role="alert" className="rounded-card border border-danger/20 bg-danger-bg/50 px-6 py-10 text-center text-sm">
        <p className="font-semibold text-ink">โหลด{label}ไม่สำเร็จ</p>
        <button
          type="button"
          onClick={query.reload}
          className="mx-auto mt-3 flex h-10 items-center gap-2 rounded-control border border-line-strong bg-white px-4 font-medium text-ink hover:bg-surface-2"
        >
          <FiRefreshCw className="size-4" />
          ลองใหม่
        </button>
      </div>
    )
  }
  if (!query.loaded) {
    return <p className="px-6 py-10 text-center text-sm text-muted">กำลังโหลด{label}…</p>
  }
  return <div className={`transition-opacity ${query.loading ? 'opacity-60' : ''}`}>{children}</div>
}

export default CustomerDetailPage
