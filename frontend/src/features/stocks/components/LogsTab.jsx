import { useEffect, useState } from 'react'
import { FiFileText, FiLock, FiScissors, FiZap } from 'react-icons/fi'
import { stockLogApi } from '../../../api/stockLogApi'
import Pagination from '../../../components/ui/Pagination'
import { formatNumber } from '../../../utils/format'
import {
  DATE_PRESETS,
  DEFAULT_LOG_FILTERS,
  LOG_TYPES,
  PAGE_SIZE_OPTIONS,
  branchColor,
  formatDate,
  formatTime,
  presetRange,
} from '../stockUtils'
import { useQuery } from '../useQuery'
import { filterSelectClass } from './styles'
import { BranchDot, EmptyState, InfoBox, LoadError, NoResults, SegmentedControl, SkeletonRows } from './ui'

const DIRECTION_OPTIONS = [
  { value: '', label: 'ทั้งหมด' },
  { value: 'in', label: 'เข้า' },
  { value: 'out', label: 'ออก' },
]

const compact = (params) => Object.fromEntries(Object.entries(params).filter(([, value]) => value !== '' && value != null))

// A transfer is stored as two rows written together: same product, opposite quantities, same moment
const TRANSFER_WINDOW_MS = 5000
const isTransferPair = (a, b) =>
  a?.log_type === 'transfer' &&
  b?.log_type === 'transfer' &&
  a.stock.id === b.stock.id &&
  a.qty_change === -b.qty_change &&
  Math.abs(new Date(a.created_at) - new Date(b.created_at)) < TRANSFER_WINDOW_MS

function LogsTab({ isAdmin, branches, stocks, filters, onFiltersChange, refreshKey, onLoaded }) {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)

  const setFilter = (patch) => {
    onFiltersChange({ ...filters, ...patch })
    setPage(1)
  }

  const range = filters.preset === 'custom' ? { from: filters.from, to: filters.to } : presetRange(filters.preset)
  const rangeInvalid = Boolean(range.from && range.to && range.from > range.to)

  const logs = useQuery(
    stockLogApi.getLogs,
    compact({
      from: range.from,
      to: range.to,
      stock_id: filters.stockId,
      log_type: filters.logType,
      direction: filters.direction,
      branch_id: isAdmin ? filters.branchId : '',
      page,
      page_size: pageSize,
    }),
    { refreshKey, enabled: !rangeInvalid },
  )

  useEffect(() => {
    if (logs.loadedAt) onLoaded(logs.loadedAt)
  }, [logs.loadedAt, onLoaded])

  const rows = logs.data ?? []
  const filtered =
    filters.preset !== DEFAULT_LOG_FILTERS.preset ||
    filters.stockId ||
    filters.logType ||
    filters.direction ||
    filters.branchId

  const clearFilters = () => {
    onFiltersChange({ ...DEFAULT_LOG_FILTERS })
    setPage(1)
  }

  return (
    <div className="space-y-5">
      <InfoBox icon={FiLock} tone="gold">
        ประวัติแก้หรือลบไม่ได้ ถ้าบันทึกผิด ให้บันทึก <span className="font-semibold">‘ปรับยอด’</span> เพื่อแก้
      </InfoBox>

      <div className="flex flex-wrap items-center gap-3">
        <SegmentedControl
          label="ช่วงวันที่"
          options={DATE_PRESETS}
          value={filters.preset}
          onChange={(preset) => {
            // Start "กำหนดเอง" from the range currently shown
            const current = presetRange(filters.preset) ?? { from: filters.from, to: filters.to }
            setFilter(preset === 'custom' ? { preset, ...current } : { preset })
          }}
        />

        {filters.preset === 'custom' && (
          <div className="flex items-center gap-2 text-sm text-muted">
            <input
              type="date"
              aria-label="ตั้งแต่วันที่"
              value={filters.from}
              max={filters.to || undefined}
              onChange={(event) => setFilter({ from: event.target.value })}
              className={filterSelectClass + ' pr-3'}
            />
            ถึง
            <input
              type="date"
              aria-label="ถึงวันที่"
              value={filters.to}
              min={filters.from || undefined}
              onChange={(event) => setFilter({ to: event.target.value })}
              className={filterSelectClass + ' pr-3'}
            />
          </div>
        )}

        <label className="flex items-center gap-2 text-sm text-muted">
          สินค้า
          <select
            value={filters.stockId}
            onChange={(event) => setFilter({ stockId: event.target.value })}
            className={`${filterSelectClass} max-w-56`}
          >
            <option value="">ทุกสินค้า</option>
            {stocks.map((stock) => (
              <option key={stock.id} value={stock.id}>
                {stock.name}
              </option>
            ))}
          </select>
        </label>

        <label className="flex items-center gap-2 text-sm text-muted">
          ประเภท
          <select value={filters.logType} onChange={(event) => setFilter({ logType: event.target.value })} className={filterSelectClass}>
            <option value="">ทุกประเภท</option>
            {Object.entries(LOG_TYPES).map(([value, type]) => (
              <option key={value} value={value}>
                {type.label}
                {type.auto ? ' (อัตโนมัติ)' : ''}
              </option>
            ))}
          </select>
        </label>

        <SegmentedControl
          label="ทิศทาง"
          options={DIRECTION_OPTIONS}
          value={filters.direction}
          onChange={(direction) => setFilter({ direction })}
        />

        {isAdmin && (
          <label className="flex items-center gap-2 text-sm text-muted">
            สาขา
            <select value={filters.branchId} onChange={(event) => setFilter({ branchId: event.target.value })} className={filterSelectClass}>
              <option value="">ทุกสาขา</option>
              {branches.map((branch) => (
                <option key={branch.id} value={branch.id}>
                  {branch.name}
                </option>
              ))}
            </select>
          </label>
        )}

        {filtered && (
          <button
            type="button"
            onClick={clearFilters}
            className="h-10 rounded-control px-2 text-sm font-semibold text-gold-dark hover:underline focus-visible:ring-3 focus-visible:ring-chart/20 focus-visible:outline-none"
          >
            ล้างตัวกรอง
          </button>
        )}
      </div>

      <section aria-busy={logs.loading}>
        {rangeInvalid ? (
          <EmptyState title="ช่วงวันที่ไม่ถูกต้อง" detail="วันที่เริ่มต้องไม่อยู่หลังวันที่สิ้นสุด" />
        ) : logs.error && !logs.loaded ? (
          <LoadError error={logs.error} onRetry={logs.reload} />
        ) : !logs.loaded ? (
          <SkeletonRows rows={8} />
        ) : rows.length === 0 ? (
          <NoResults onClear={filtered ? clearFilters : undefined} />
        ) : (
          <div className={`overflow-x-auto rounded-card border border-line transition-opacity ${logs.loading ? 'opacity-60' : ''}`}>
            <table className="w-full min-w-[1180px] table-fixed text-sm">
              <colgroup>
                <col className="w-[130px]" />
                <col />
                <col className="w-[96px]" />
                <col className="w-[180px]" />
                <col className="w-[96px]" />
                <col className="w-[150px]" />
                <col className="w-[110px]" />
                <col className="w-[230px]" />
              </colgroup>
              <thead className="text-xs text-muted">
                <tr className="border-b border-line text-left whitespace-nowrap">
                  <th className="py-3.5 pr-3.5 pl-5 font-medium">วันเวลา</th>
                  <th className="px-3.5 py-3.5 font-medium">สินค้า</th>
                  <th className="px-3.5 py-3.5 font-medium">สาขา</th>
                  <th className="px-3.5 py-3.5 font-medium">ประเภท</th>
                  <th className="px-3.5 py-3.5 text-right font-medium">จำนวน</th>
                  <th className="px-3.5 py-3.5 font-medium">อ้างอิง</th>
                  <th className="px-3.5 py-3.5 font-medium">พนักงาน</th>
                  <th className="py-3.5 pr-5 pl-3.5 font-medium">หมายเหตุ</th>
                </tr>
              </thead>
              <tbody>
                {rows.map((log, index) => {
                  const pairedWithNext = isTransferPair(log, rows[index + 1])
                  const pairedWithPrevious = isTransferPair(rows[index - 1], log)
                  const paired = pairedWithNext || pairedWithPrevious
                  return (
                    <tr
                      key={log.id}
                      className={`${paired ? 'bg-[#F3F8FB]' : 'bg-white'} ${
                        pairedWithNext ? 'border-b border-dashed border-[#C9DCE6]' : 'border-b border-line last:border-0'
                      } transition-[filter] hover:brightness-[0.985]`}
                    >
                      <td className="py-3.5 pr-3.5 pl-5">
                        <p className="font-medium text-ink">{formatDate(log.created_at)}</p>
                        <p className="mt-0.5 text-xs text-muted">{formatTime(log.created_at)}</p>
                      </td>
                      <td className="px-3.5 py-3.5 text-ink [overflow-wrap:anywhere]">{log.stock.name}</td>
                      <td className="px-3.5 py-3.5">
                        <span className="flex items-center gap-2 font-mono text-xs text-ink" title={log.branch.name}>
                          <BranchDot color={branchColor(branches, log.branch.id)} />
                          {log.branch.code}
                        </span>
                      </td>
                      <td className="px-3.5 py-3.5">
                        <LogTypeLabel type={log.log_type} />
                      </td>
                      <td
                        className={`px-3.5 py-3.5 text-right font-semibold tabular-nums ${
                          log.qty_change > 0 ? 'text-success' : 'text-danger'
                        }`}
                      >
                        {log.qty_change > 0 ? '+' : '−'}
                        {formatNumber(Math.abs(log.qty_change))}
                      </td>
                      <td className="px-3.5 py-3.5 text-muted">
                        <Reference log={log} />
                      </td>
                      <td className="px-3.5 py-3.5 text-ink [overflow-wrap:anywhere]">{log.employee.name || '—'}</td>
                      <td className="py-3.5 pr-5 pl-3.5 text-muted [overflow-wrap:anywhere]">{log.note || '—'}</td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
      </section>

      {logs.meta && logs.meta.total > 0 && !rangeInvalid && !(logs.error && !logs.loaded) && (
        <Pagination
          meta={logs.meta}
          unit="รายการ"
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

function LogTypeLabel({ type }) {
  const meta = LOG_TYPES[type] ?? { label: type, color: '#8A8C95' }
  return (
    <span className="flex flex-wrap items-center gap-x-2 gap-y-0.5">
      <span className="flex items-center gap-2 text-ink">
        <span aria-hidden="true" className="size-2 rounded-full" style={{ backgroundColor: meta.color }} />
        {meta.label}
      </span>
      {meta.auto && (
        <span className="flex items-center gap-0.5 text-[11px] text-subtle">
          <FiZap className="size-3" />
          อัตโนมัติ
        </span>
      )}
    </span>
  )
}

function Reference({ log }) {
  if (log.order_item_id) {
    return (
      <span className="flex items-center gap-1.5">
        <FiFileText className="size-3.5" />
        ออเดอร์
      </span>
    )
  }
  if (log.usage_id) {
    return (
      <span className="flex items-center gap-1.5">
        <FiScissors className="size-3.5" />
        การใช้บริการ
      </span>
    )
  }
  return '—'
}

export default LogsTab
