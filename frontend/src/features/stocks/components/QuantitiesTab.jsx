import { useEffect, useState } from 'react'
import { FiArchive, FiArrowRight, FiClock, FiDownload, FiRepeat, FiSearch, FiSliders, FiUpload } from 'react-icons/fi'
import { stockLogApi } from '../../../api/stockLogApi'
import Pagination from '../../../components/ui/Pagination'
import Toggle from '../../../components/ui/Toggle'
import { formatNumber } from '../../../utils/format'
import { DEFAULT_QTY_FILTERS, MAX_PAGE_SIZE, PAGE_SIZE_OPTIONS, branchColor } from '../stockUtils'
import { useQuery } from '../useQuery'
import RowMenu from './RowMenu'
import { filterSelectClass } from './styles'
import { BranchDot, CategoryText, EmptyState, LoadError, NoResults, SegmentedControl, SkeletonRows } from './ui'

const CATEGORY_OPTIONS = [
  { value: '', label: 'ทั้งหมด' },
  { value: 'premium', label: 'Premium' },
  { value: 'normal', label: 'Normal' },
]

const LOW_BG = 'bg-[#FFF5D6]'
const OUT_BG = 'bg-[#FDEEEE]'
const rowTone = (row) => (row.quantity === 0 ? OUT_BG : row.is_low_stock ? LOW_BG : 'bg-white')

// Drop empty values so the URL only carries real filters
const compact = (params) => Object.fromEntries(Object.entries(params).filter(([, value]) => value !== '' && value != null))

// Consecutive rows of the same product share one product cell (rowSpan)
function groupRows(rows) {
  const groups = []
  for (const row of rows) {
    const last = groups.at(-1)
    if (last && last.stockId === row.stock_id) last.rows.push(row)
    else groups.push({ stockId: row.stock_id, rows: [row] })
  }
  return groups
}

function QuantitiesTab({
  isAdmin,
  branches,
  ownBranch,
  activeStockCount,
  filters,
  onFiltersChange,
  refreshKey,
  onLoaded,
  onAction,
}) {
  const [searchInput, setSearchInput] = useState(filters.search)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(PAGE_SIZE_OPTIONS[0])
  const branchId = isAdmin ? filters.branchId : ''

  const setFilter = (patch) => {
    onFiltersChange({ ...filters, ...patch })
    setPage(1)
  }

  useEffect(() => {
    const timer = setTimeout(() => {
      const search = searchInput.trim()
      if (search !== filters.search) {
        onFiltersChange({ ...filters, search })
        setPage(1)
      }
    }, 300)
    return () => clearTimeout(timer)
  }, [searchInput, filters, onFiltersChange])

  const list = useQuery(
    stockLogApi.getQuantities,
    compact({
      branch_id: branchId,
      search: filters.search,
      category: filters.category,
      active: filters.active,
      low_stock: filters.lowStock ? 'true' : '',
      page,
      page_size: pageSize,
    }),
    { refreshKey },
  )

  // Every zero-stock row is also a low-stock row (min_qty ≥ 0), so one request feeds both summary numbers
  const low = useQuery(
    stockLogApi.getQuantities,
    compact({ branch_id: branchId, low_stock: 'true', page_size: MAX_PAGE_SIZE }),
    { refreshKey },
  )

  useEffect(() => {
    if (list.loadedAt) onLoaded(list.loadedAt)
  }, [list.loadedAt, onLoaded])

  const lowCount = low.meta?.total
  const outRows = (low.data ?? []).filter((row) => row.quantity === 0)
  const showBranch = isAdmin
  const rows = list.data ?? []
  const filtered = filters.search || filters.category || filters.active !== 'true' || filters.lowStock || filters.branchId

  const clearFilters = () => {
    setSearchInput('')
    onFiltersChange({ ...DEFAULT_QTY_FILTERS })
    setPage(1)
  }

  return (
    <div className="space-y-5">
      <dl className="flex flex-wrap items-stretch rounded-card border border-line bg-white">
        <SummaryCell label="สินค้าที่ใช้งานอยู่" value={activeStockCount} />
        <SummaryCell label="รายการใกล้หมด" value={lowCount} unit="รายการ" tone="text-warn" />
        <div className="flex min-w-64 flex-1 flex-wrap items-center justify-end gap-x-6 gap-y-2 px-5 py-4">
          <p className="text-sm text-muted">
            หมดแล้ว{' '}
            {outRows.length === 0 ? (
              <span className="font-semibold text-danger">—</span>
            ) : (
              <span className="font-semibold text-danger">
                {outRows.map((row) => (isAdmin ? `${row.stock_name} (${row.branch_name})` : row.stock_name)).join(', ')}
              </span>
            )}
          </p>
          <button
            type="button"
            onClick={() => setFilter({ lowStock: true })}
            className="flex items-center gap-1.5 rounded-sm text-sm font-semibold text-ink hover:underline focus-visible:ring-3 focus-visible:ring-chart/20 focus-visible:outline-none"
          >
            ดูเฉพาะที่ใกล้หมด
            <FiArrowRight className="size-4" />
          </button>
        </div>
      </dl>

      <div className="flex flex-wrap items-center gap-3">
        <label className="relative w-full max-w-72">
          <FiSearch className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-subtle" />
          <input
            type="search"
            value={searchInput}
            onChange={(event) => setSearchInput(event.target.value)}
            placeholder="ค้นหาชื่อสินค้า"
            aria-label="ค้นหาชื่อสินค้า"
            maxLength={100}
            className="h-10 w-full rounded-control border border-line-strong bg-white pr-3 pl-10 text-sm placeholder:text-subtle focus:border-chart/40 focus:ring-3 focus:ring-chart/10 focus:outline-none"
          />
        </label>

        <SegmentedControl
          label="หมวด"
          options={CATEGORY_OPTIONS}
          value={filters.category}
          onChange={(category) => setFilter({ category })}
        />

        <label className="flex items-center gap-2 text-sm text-muted">
          สถานะ
          <select value={filters.active} onChange={(event) => setFilter({ active: event.target.value })} className={filterSelectClass}>
            <option value="true">ใช้งาน</option>
            <option value="false">ปิดใช้งาน</option>
            <option value="all">ทั้งหมด</option>
          </select>
        </label>

        {isAdmin ? (
          <label className="flex items-center gap-2 text-sm text-muted">
            สาขา
            <select
              value={filters.branchId}
              onChange={(event) => setFilter({ branchId: event.target.value })}
              className={filterSelectClass}
            >
              <option value="">ทุกสาขา</option>
              {branches.map((branch) => (
                <option key={branch.id} value={branch.id}>
                  {branch.name}
                </option>
              ))}
            </select>
          </label>
        ) : (
          <OwnBranchChip branch={ownBranch} branches={branches} />
        )}

        <div className="flex h-10 items-center rounded-control border border-line-strong bg-white px-3">
          <Toggle checked={filters.lowStock} onChange={(lowStock) => setFilter({ lowStock })} label="เฉพาะที่ใกล้หมด">
            <span className="whitespace-nowrap">เฉพาะที่ใกล้หมด</span>
          </Toggle>
        </div>
      </div>

      <section aria-busy={list.loading}>
        {list.error && !list.loaded ? (
          <LoadError error={list.error} onRetry={list.reload} />
        ) : !list.loaded ? (
          <SkeletonRows />
        ) : rows.length === 0 ? (
          filtered ? (
            <NoResults onClear={clearFilters} />
          ) : (
            <EmptyState title="ยังไม่มีสินค้าในสต็อก" detail={isAdmin ? 'เพิ่มสินค้าได้ที่แท็บ “รายการสินค้า”' : undefined} />
          )
        ) : (
          <div className={`overflow-x-auto rounded-card border border-line transition-opacity ${list.loading ? 'opacity-60' : ''}`}>
            <table className="w-full min-w-[760px] table-fixed text-sm">
              {/* Percent widths spread the short columns across wide screens instead of piling them up on the right */}
              <colgroup>
                <col />
                {showBranch && <col className="w-[20%]" />}
                <col className={showBranch ? 'w-[16%]' : 'w-[20%]'} />
                <col className={showBranch ? 'w-[18%]' : 'w-[22%]'} />
                <col className="w-[76px]" />
              </colgroup>
              <thead className="text-xs text-muted">
                <tr className="border-b border-line text-left whitespace-nowrap">
                  <th className="py-3.5 pr-3.5 pl-5 font-medium">สินค้า</th>
                  {showBranch && <th className="px-3.5 py-3.5 font-medium">สาขา</th>}
                  <th className="px-3.5 py-3.5 font-medium">คงเหลือ*</th>
                  <th className="px-3.5 py-3.5 font-medium">สถานะ</th>
                  <th className="py-3.5 pr-5 pl-3.5">
                    <span className="sr-only">การทำงาน</span>
                  </th>
                </tr>
              </thead>
              <tbody>
                {groupRows(rows).map((group) =>
                  group.rows.map((row, index) => (
                    <tr
                      key={`${row.stock_id}-${row.branch_id}`}
                      className={`${rowTone(row)} ${
                        index === group.rows.length - 1 ? 'border-b border-line last:border-0' : ''
                      } transition-[filter] hover:brightness-[0.985]`}
                    >
                      {index === 0 && (
                        <td rowSpan={group.rows.length} className="border-r-0 bg-white py-3.5 pr-3.5 pl-5 align-top">
                          <p className="font-semibold text-ink [overflow-wrap:anywhere]">{row.stock_name}</p>
                          <p className="mt-1 flex flex-wrap items-center gap-x-2.5 text-xs text-muted">
                            <CategoryText category={row.category} />
                            {row.uses_per_unit > 1 && <span>1 ชิ้นใช้ได้ {formatNumber(row.uses_per_unit)} ครั้ง</span>}
                          </p>
                        </td>
                      )}
                      {showBranch && (
                        <td className="px-3.5 py-3.5">
                          <span className="flex items-center gap-2 text-ink">
                            <BranchDot color={branchColor(branches, row.branch_id)} />
                            {row.branch_name}
                          </span>
                        </td>
                      )}
                      <td
                        className={`px-3.5 py-3.5 text-[15px] font-semibold tabular-nums ${
                          row.quantity === 0 ? 'text-danger' : 'text-ink'
                        }`}
                      >
                        {/* Left-anchored like the other columns; the fixed-width box keeps digits right-aligned */}
                        <span className="inline-block min-w-[5ch] text-right">{formatNumber(row.quantity)}</span>
                      </td>
                      <td className="px-3.5 py-3.5">
                        {row.is_active ? (
                          <span className="text-ink">ใช้งาน</span>
                        ) : (
                          <span className="text-subtle italic">ปิดใช้งาน</span>
                        )}
                      </td>
                      <td className="py-2 pr-5 pl-3.5">
                        <div className="flex justify-end">
                          <RowMenu label={`การทำงานของ ${row.stock_name} ${row.branch_name}`} items={menuItems(row, onAction)} />
                        </div>
                      </td>
                    </tr>
                  )),
                )}
              </tbody>
            </table>
          </div>
        )}
      </section>

      {list.meta && list.meta.total > 0 && !(list.error && !list.loaded) && (
        <div className="space-y-3">
          <div className="flex flex-wrap items-center gap-x-5 gap-y-2 text-xs text-muted">
            <span>* ยอดนับเฉพาะชิ้นที่ยังไม่เปิด (กระปุกที่เปิดใช้แล้วไม่นับ)</span>
            <span className="flex items-center gap-1.5">
              <span className={`size-3.5 rounded border border-[#F2DFA6] ${LOW_BG}`} />
              ใกล้หมด
            </span>
            <span className="flex items-center gap-1.5">
              <span className={`size-3.5 rounded border border-[#F3CFCF] ${OUT_BG}`} />
              หมดแล้ว
            </span>
          </div>
          <Pagination
            meta={list.meta}
            unit="รายการ"
            pageSizeOptions={PAGE_SIZE_OPTIONS}
            onPageChange={setPage}
            onPageSizeChange={(size) => {
              setPageSize(size)
              setPage(1)
            }}
          />
        </div>
      )}
    </div>
  )
}

function menuItems(row, onAction) {
  const items = [
    { label: 'รับเข้า', icon: FiDownload, onSelect: () => onAction('movement', row, 'receive') },
    { label: 'บันทึกออก', icon: FiUpload, onSelect: () => onAction('movement', row, 'expired') },
    { label: 'ปรับยอด', icon: FiSliders, onSelect: () => onAction('movement', row, 'adjust') },
    { label: 'โอน', icon: FiRepeat, onSelect: () => onAction('transfer', row) },
    'divider',
    { label: 'ดูประวัติ', icon: FiClock, onSelect: () => onAction('history', row) },
  ]
  if (row.uses_per_unit > 1) {
    items.push({ label: 'ดูกระปุกที่เปิดอยู่', icon: FiArchive, onSelect: () => onAction('containers', row) })
  }
  return items
}

function SummaryCell({ label, value, unit, tone = 'text-ink' }) {
  return (
    <div className="min-w-44 border-r border-line px-5 py-4">
      <dt className="text-xs text-muted">{label}</dt>
      <dd className="mt-2 flex items-baseline gap-1.5">
        <span className={`text-[26px] leading-none font-bold tabular-nums ${tone}`}>
          {value == null ? '–' : formatNumber(value)}
        </span>
        {unit && <span className="text-sm text-muted">{unit}</span>}
      </dd>
    </div>
  )
}

export function OwnBranchChip({ branch, branches }) {
  return (
    <span className="flex h-10 items-center gap-2 rounded-full bg-[#EEF1F8] px-3.5 text-sm font-semibold text-chart">
      <BranchDot color={branchColor(branches, branch?.id)} />
      {branch?.name ?? 'สาขาของฉัน'}
    </span>
  )
}

export default QuantitiesTab
