import { useEffect, useState } from 'react'
import { FiPlus, FiRefreshCw, FiSearch, FiWifiOff, FiX } from 'react-icons/fi'
import { customerApi } from '../../api/customerApi'
import Pagination from '../../components/ui/Pagination'
import { useQuery } from '../../hooks/useQuery'
import CustomerFormModal from './components/CustomerFormModal'
import CustomerTable from './components/CustomerTable'
import { PAGE_SIZE_OPTIONS, TIER_FILTERS } from './customerUtils'

function CustomersPage() {
  const [searchInput, setSearchInput] = useState('')
  const [search, setSearch] = useState('')
  const [tier, setTier] = useState('')
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(PAGE_SIZE_OPTIONS[1])
  const [form, setForm] = useState(null) // { customer } — customer null = create

  // Empty q / tier are left out (undefined) so the request has no ?q= or ?tier=
  const list = useQuery(customerApi.search, {
    q: search || undefined,
    tier: tier || undefined,
    page,
    page_size: pageSize,
  })
  const customers = list.data ?? []
  const filtered = search !== '' || tier !== ''

  useEffect(() => {
    const timer = setTimeout(() => {
      setSearch(searchInput.trim())
      setPage(1)
    }, 300)
    return () => clearTimeout(timer)
  }, [searchInput])

  const pickTier = (value) => {
    setTier(value)
    setPage(1)
  }

  const clearFilters = () => {
    setSearchInput('')
    setSearch('')
    setTier('')
    setPage(1)
  }

  return (
    <div className="mx-auto max-w-[1400px] space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-[28px] font-bold tracking-tight text-ink">ลูกค้า</h1>
          <p className="mt-1 text-sm text-muted">ค้นหาลูกค้าจากชื่อ เลข HN หรือเบอร์โทร ใช้ร่วมกันทุกสาขา</p>
        </div>
        <button
          type="button"
          onClick={() => setForm({ customer: null })}
          className="flex h-11 items-center gap-2 rounded-control bg-primary px-4 text-sm font-semibold text-white shadow-sm transition hover:bg-[#4d1621]"
        >
          <FiPlus className="size-4" />
          เพิ่มลูกค้า
        </button>
      </div>

      <div className="flex flex-wrap items-center gap-3">
        <label className="relative w-full max-w-sm">
          <FiSearch className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-subtle" />
          <input
            type="search"
            value={searchInput}
            onChange={(event) => setSearchInput(event.target.value)}
            placeholder="ค้นหาชื่อ HN หรือเบอร์โทร"
            aria-label="ค้นหาลูกค้า"
            className="h-10 w-full rounded-control border border-line-strong bg-white pr-3 pl-10 text-sm placeholder:text-subtle focus:border-chart/40 focus:ring-3 focus:ring-chart/10 focus:outline-none"
          />
        </label>

        <div role="group" aria-label="ระดับสมาชิก" className="flex rounded-[12px] border border-line bg-surface-2 p-1">
          {TIER_FILTERS.map((option) => (
            <button
              key={option.value || 'all'}
              type="button"
              aria-pressed={tier === option.value}
              onClick={() => pickTier(option.value)}
              className={`h-9 rounded-lg px-3.5 text-sm transition focus-visible:ring-3 focus-visible:ring-chart/20 focus-visible:outline-none ${
                tier === option.value ? 'bg-white font-semibold text-ink shadow-sm' : 'text-muted hover:text-ink'
              }`}
            >
              {option.label}
            </button>
          ))}
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
          <EmptyState title="กำลังโหลดรายชื่อลูกค้า…" />
        ) : customers.length === 0 ? (
          <EmptyState
            title={filtered ? 'ไม่พบลูกค้าที่ตรงกับตัวกรอง' : 'ยังไม่มีลูกค้า'}
            detail={filtered ? 'ลองค้นหาด้วยเบอร์โทรหรือเลข HN หรือล้างตัวกรอง' : undefined}
          />
        ) : (
          <CustomerTable customers={customers} onEdit={(customer) => setForm({ customer })} />
        )}
      </section>

      {list.meta && list.meta.total > 0 && !list.error && (
        <Pagination
          meta={list.meta}
          unit="คน"
          pageSizeOptions={PAGE_SIZE_OPTIONS}
          onPageChange={setPage}
          onPageSizeChange={(size) => {
            setPageSize(size)
            setPage(1)
          }}
        />
      )}

      {form && (
        <CustomerFormModal
          customer={form.customer}
          onClose={() => setForm(null)}
          onSaved={() => {
            setForm(null)
            list.reload()
          }}
        />
      )}
    </div>
  )
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
      <p className="mt-3 font-semibold text-ink">โหลดรายชื่อลูกค้าไม่สำเร็จ</p>
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

export default CustomersPage
