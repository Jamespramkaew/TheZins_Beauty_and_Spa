import { useId, useState } from 'react'
import { FiArchive, FiX } from 'react-icons/fi'
import Modal from '../../../components/Modal'
import { stockApi } from '../../../api/stockApi'
import { formatNumber } from '../../../utils/format'
import { CLOSE_REASONS, MAX_PAGE_SIZE, branchColor, daysSince, formatDate } from '../stockUtils'
import { useQuery } from '../useQuery'
import { filterSelectClass } from './styles'
import { BranchDot, LoadError, SegmentedControl } from './ui'

const fetchContainers = ({ stockId, ...params }) => stockApi.getContainers(stockId, params)

// stock: { id, name, uses_per_unit } · branchId: initial branch filter (admin only)
function ContainersDrawer({ isAdmin, branches, stock, branchId: initialBranchId, onClose }) {
  const titleId = useId()
  const [status, setStatus] = useState('open')
  const [branchId, setBranchId] = useState(isAdmin ? (initialBranchId ?? '') : '')
  const scope = { stockId: stock.id, ...(isAdmin && branchId && { branch_id: branchId }) }

  const list = useQuery(fetchContainers, { ...scope, status, page_size: MAX_PAGE_SIZE })
  const openCount = useQuery(fetchContainers, { ...scope, status: 'open', page_size: 1 }).meta?.total
  const closedCount = useQuery(fetchContainers, { ...scope, status: 'closed', page_size: 1 }).meta?.total

  const tabs = [
    { value: 'open', label: `เปิดอยู่${openCount == null ? '' : ` (${openCount})`}` },
    { value: 'closed', label: `ปิดแล้ว${closedCount == null ? '' : ` (${closedCount})`}` },
    { value: 'all', label: 'ทั้งหมด' },
  ]
  const containers = list.data ?? []

  return (
    <Modal placement="right" onClose={onClose} labelledBy={titleId} className="sm:max-w-[460px]">
      <header className="flex items-start justify-between gap-4 border-b border-line px-5 py-4 sm:px-6">
        <div className="min-w-0">
          <p className="text-xs font-semibold text-gold-dark">กระปุกที่เปิดใช้อยู่</p>
          <h2 id={titleId} className="mt-1 text-lg font-bold text-ink [overflow-wrap:anywhere]">
            {stock.name}
          </h2>
          <p className="mt-0.5 text-[13px] text-muted">1 กระปุกใช้ได้ {formatNumber(stock.uses_per_unit)} ครั้ง</p>
        </div>
        <button
          type="button"
          onClick={onClose}
          aria-label="ปิด"
          className="flex size-10 shrink-0 items-center justify-center rounded-control bg-surface-2 text-muted transition hover:text-ink focus-visible:ring-3 focus-visible:ring-chart/20 focus-visible:outline-none"
        >
          <FiX className="size-4" />
        </button>
      </header>

      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-line px-5 py-3.5 sm:px-6">
        <SegmentedControl label="สถานะกระปุก" options={tabs} value={status} onChange={setStatus} size="sm" />
        {isAdmin && (
          <select
            aria-label="สาขา"
            value={branchId}
            onChange={(event) => setBranchId(event.target.value)}
            className={filterSelectClass}
          >
            <option value="">ทุกสาขา</option>
            {branches.map((branch) => (
              <option key={branch.id} value={branch.id}>
                {branch.name}
              </option>
            ))}
          </select>
        )}
      </div>

      <div aria-busy={list.loading} className="min-h-0 flex-1 overflow-y-auto px-5 py-4 sm:px-6">
        {list.error && !list.loaded ? (
          <LoadError error={list.error} onRetry={list.reload} />
        ) : !list.loaded ? (
          <div className="space-y-3">
            {[0, 1, 2].map((index) => (
              <div key={index} className="h-20 animate-pulse rounded-card bg-surface-2" />
            ))}
          </div>
        ) : containers.length === 0 ? (
          <div className="flex flex-col items-center py-16 text-center">
            <span className="flex size-11 items-center justify-center rounded-full bg-surface-2 text-subtle">
              <FiArchive className="size-5" />
            </span>
            <p className="mt-3 font-semibold text-ink">
              {status === 'closed' ? 'ยังไม่มีกระปุกที่ปิดแล้ว' : status === 'open' ? 'ไม่มีกระปุกที่เปิดอยู่' : 'ยังไม่มีกระปุก'}
            </p>
          </div>
        ) : (
          <ul className={`space-y-3 transition-opacity ${list.loading ? 'opacity-60' : ''}`}>
            {containers.map((container) => (
              <ContainerCard
                key={container.id}
                container={container}
                usesPerUnit={stock.uses_per_unit}
                color={branchColor(branches, container.branch_id)}
              />
            ))}
          </ul>
        )}
        {list.meta && list.meta.total > containers.length && (
          <p className="mt-4 text-center text-xs text-muted">
            แสดง {formatNumber(containers.length)} จาก {formatNumber(list.meta.total)} กระปุกล่าสุด
          </p>
        )}
      </div>
    </Modal>
  )
}

function ContainerCard({ container, usesPerUnit, color }) {
  const closed = container.closed_at != null
  const reason = CLOSE_REASONS[container.close_reason]
  const used = container.uses_count ?? 0
  const percent = usesPerUnit > 0 ? Math.min(100, Math.round((used / usesPerUnit) * 100)) : 0

  return (
    <li className="rounded-card border border-line bg-white p-4">
      <div className="flex items-center justify-between gap-3">
        <span className="flex items-center gap-2 font-semibold text-ink">
          <BranchDot color={color} />
          {container.branch?.name ?? '—'}
        </span>
        {closed ? (
          <span className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${reason?.className ?? 'bg-surface-2 text-muted'}`}>
            {reason?.label ?? 'ปิดแล้ว'}
          </span>
        ) : (
          <span className="rounded-full bg-[#EEF1F8] px-2.5 py-0.5 text-xs font-medium text-chart">เปิดอยู่</span>
        )}
      </div>

      <p className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-sm text-muted">
        <span>เปิดเมื่อ {formatDate(container.opened_at)}</span>
        {closed ? (
          <span>ปิดเมื่อ {formatDate(container.closed_at)}</span>
        ) : (
          <span className="font-semibold text-chart">เปิดมา {formatNumber(daysSince(container.opened_at))} วัน</span>
        )}
      </p>

      {closed && (
        <div className="mt-3">
          <div className="h-1.5 overflow-hidden rounded-full bg-surface-2">
            <div className="h-full rounded-full bg-chart" style={{ width: `${percent}%` }} />
          </div>
          <p className="mt-1.5 text-xs text-muted tabular-nums">
            ใช้ไป {formatNumber(used)}/{formatNumber(usesPerUnit)} ครั้ง
          </p>
        </div>
      )}
    </li>
  )
}

export default ContainersDrawer
