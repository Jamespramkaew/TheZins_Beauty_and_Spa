import { useState } from 'react'
import { FiRefreshCw, FiWifiOff, FiX } from 'react-icons/fi'
import { orderApi } from '../../api/orderApi'
import FilterGroup from '../../components/ui/FilterGroup'
import Pagination from '../../components/ui/Pagination'
import { useAuth } from '../../hooks/useAuth'
import { useQuery } from '../../hooks/useQuery'
import OrderTable from './components/OrderTable'
import { PAGE_SIZE_OPTIONS, STATUS_FILTERS } from './orderUtils'

function OrdersPage() {
  const { user } = useAuth()
  const isAdmin = user.role === 'admin'
  const [status, setStatus] = useState('')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(PAGE_SIZE_OPTIONS[1])

  // Empty filters are left out (undefined) so the request has no ?status= / ?from= / ?to=
  const dateParams = { from: from || undefined, to: to || undefined }
  const list = useQuery(orderApi.getAll, { ...dateParams, status: status || undefined, page, page_size: pageSize })
  const orders = list.data ?? []
  const filtered = status !== '' || from !== '' || to !== ''

  // Counts beside the status buttons follow the date range (the number = what you get when you press it).
  // Only meta.total is needed, so each asks for 1 row. ชำระครบ = ทั้งหมด − ยังไม่ชำระ − บางส่วน.
  const allCount = useOrderTotal(dateParams)
  const unpaidCount = useOrderTotal({ ...dateParams, status: 'unpaid' })
  const partialCount = useOrderTotal({ ...dateParams, status: 'partial' })
  const counts = {
    '': allCount,
    unpaid: unpaidCount,
    partial: partialCount,
    paid: allCount != null && unpaidCount != null && partialCount != null ? allCount - unpaidCount - partialCount : undefined,
  }

  const pickStatus = (value) => {
    setStatus(value)
    setPage(1)
  }

  const pickDate = (setter) => (event) => {
    setter(event.target.value)
    setPage(1)
  }

  const clearFilters = () => {
    setStatus('')
    setFrom('')
    setTo('')
    setPage(1)
  }

  return (
    <div className="mx-auto max-w-[1400px] space-y-6">
      <div>
        <h1 className="text-[28px] font-bold tracking-tight text-ink">บิลขาย</h1>
        <p className="mt-1 text-sm text-muted">
          {isAdmin ? 'บิลของทุกสาขา เรียงจากวันที่ขายล่าสุด' : 'บิลของสาขานี้ เรียงจากวันที่ขายล่าสุด'}
        </p>
      </div>

      <div className="flex flex-wrap items-center gap-3">
        <FilterGroup label="สถานะการชำระ" options={STATUS_FILTERS} value={status} counts={counts} onChange={pickStatus} />

        <div className="flex flex-wrap items-center gap-2 text-sm text-muted">
          <label className="flex items-center gap-2">
            ตั้งแต่
            <input
              type="date"
              value={from}
              max={to || undefined}
              onChange={pickDate(setFrom)}
              className="h-10 rounded-control border border-line-strong bg-white px-3 text-sm text-ink focus:border-chart/40 focus:ring-3 focus:ring-chart/10 focus:outline-none"
            />
          </label>
          <label className="flex items-center gap-2">
            ถึง
            <input
              type="date"
              value={to}
              min={from || undefined}
              onChange={pickDate(setTo)}
              className="h-10 rounded-control border border-line-strong bg-white px-3 text-sm text-ink focus:border-chart/40 focus:ring-3 focus:ring-chart/10 focus:outline-none"
            />
          </label>
        </div>

        {filtered && (
          <button
            type="button"
            onClick={clearFilters}
            className="flex h-10 items-center gap-1.5 rounded-control px-3 text-sm text-muted transition hover:bg-surface-2 hover:text-ink"
          >
            <FiX className="size-4" />
            ล้างตัวกรอง
          </button>
        )}
      </div>

      <section aria-busy={list.loading} className={`transition-opacity ${list.loading && list.loaded ? 'opacity-60' : ''}`}>
        {list.error ? (
          <LoadError onRetry={list.reload} />
        ) : !list.loaded ? (
          <EmptyState title="กำลังโหลดบิล…" />
        ) : orders.length === 0 ? (
          <EmptyState
            title={filtered ? 'ไม่พบบิลที่ตรงกับตัวกรอง' : 'ยังไม่มีบิล'}
            detail={filtered ? 'ลองเปลี่ยนช่วงวันที่ หรือล้างตัวกรอง' : undefined}
          />
        ) : (
          <OrderTable orders={orders} />
        )}
      </section>

      {list.meta && list.meta.total > 0 && !list.error && (
        <Pagination
          meta={list.meta}
          unit="บิล"
          pageSizeOptions={PAGE_SIZE_OPTIONS}
          onPageChange={setPage}
          onPageSizeChange={(size) => {
            setPageSize(size)
            setPage(1)
          }}
        />
      )}
    </div>
  )
}

// meta.total of GET /orders for these filters (1 row requested — only the count is used)
function useOrderTotal(params) {
  return useQuery(orderApi.getAll, { ...params, page: 1, page_size: 1 }).meta?.total
}

function EmptyState({ title, detail }) {
  return (
    <div className="flex flex-col items-center justify-center rounded-card border border-dashed border-line-strong bg-surface px-6 py-16 text-center">
      <p className="font-semibold text-ink">{title}</p>
      {detail && <p className="mt-1 text-sm text-muted">{detail}</p>}
    </div>
  )
}

function LoadError({ onRetry }) {
  return (
    <div
      role="alert"
      className="flex flex-col items-center justify-center rounded-card border border-danger/20 bg-danger-bg/50 px-6 py-14 text-center"
    >
      <span className="flex size-11 items-center justify-center rounded-full bg-white text-danger shadow-sm">
        <FiWifiOff className="size-5" />
      </span>
      <p className="mt-3 font-semibold text-ink">โหลดบิลไม่สำเร็จ</p>
      <p className="mt-1 text-sm text-muted">ตรวจสอบการเชื่อมต่อ แล้วลองใหม่อีกครั้ง</p>
      <button
        type="button"
        onClick={onRetry}
        className="mt-4 flex h-10 items-center gap-2 rounded-control border border-line-strong bg-white px-4 text-sm font-medium text-ink hover:bg-surface-2"
      >
        <FiRefreshCw className="size-4" />
        ลองใหม่
      </button>
    </div>
  )
}

export default OrdersPage
