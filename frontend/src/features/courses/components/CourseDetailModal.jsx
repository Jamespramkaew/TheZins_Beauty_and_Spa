import { useId } from 'react'
import { FiEdit2, FiX } from 'react-icons/fi'
import Modal from '../../../components/Modal'
import { formatBaht, formatNumber } from '../../../utils/format'
import { formatValidDay, pricePerSession, toNumber } from '../courseUtils'
import CategoryBadge from './CategoryBadge'

// Read-only view — the table collapses items into "+N", this shows the full list
function CourseDetailModal({ course, canManage, onClose, onEdit }) {
  const titleId = useId()

  return (
    <Modal onClose={onClose} labelledBy={titleId} className="max-w-[560px]">
      <header className="flex items-start justify-between gap-4 border-b border-line px-5 py-4 sm:px-6">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <CategoryBadge category={course.category} />
            <span
              className={`inline-flex rounded-full px-2.5 py-1 text-xs font-medium ${
                course.is_active ? 'bg-success-bg text-success' : 'bg-surface-2 text-muted'
              }`}
            >
              {course.is_active ? 'เปิดขาย' : 'ปิดใช้งาน'}
            </span>
          </div>
          <h2 id={titleId} className="mt-2 text-lg font-bold text-ink [overflow-wrap:anywhere]">
            {course.name}
          </h2>
        </div>
        <button
          type="button"
          onClick={onClose}
          aria-label="ปิด"
          className="flex size-9 shrink-0 items-center justify-center rounded-control bg-surface-2 text-muted transition hover:text-ink"
        >
          <FiX className="size-4" />
        </button>
      </header>

      <div className="min-h-0 flex-1 overflow-y-auto px-5 py-5 sm:px-6">
        <p className="flex flex-wrap items-baseline gap-x-2">
          <span className="text-[28px] leading-none font-bold tracking-tight text-ink">
            {formatBaht(toNumber(course.price))}
          </span>
          <span className="text-sm text-muted">/ {formatNumber(course.max_session)} ครั้ง</span>
        </p>

        <dl className="mt-4 grid grid-cols-3 gap-2">
          <Stat label="เฉลี่ย/ครั้ง" value={formatBaht(pricePerSession(course.price, course.max_session))} />
          <Stat label="อายุคอร์ส" value={formatValidDay(course.valid_day)} />
          <Stat label="ค่ามือ/ครั้ง" value={formatBaht(toNumber(course.service_comm))} />
        </dl>

        <section className="mt-6">
          <h3 className="mb-2 flex items-baseline justify-between text-sm font-semibold text-ink">
            สินค้าที่ใช้ต่อบริการ
            <span className="text-xs font-normal text-muted">{formatNumber(course.items.length)} รายการ</span>
          </h3>
          {course.items.length === 0 ? (
            <p className="rounded-control bg-surface-2 px-4 py-3 text-[13px] text-subtle italic">ไม่ใช้สินค้า</p>
          ) : (
            // Order comes from the API — never re-sort here
            <ol className="divide-y divide-line rounded-control border border-line">
              {course.items.map((item, index) => (
                <li key={item.stock.id} className="flex items-center gap-3 px-4 py-2.5 text-sm">
                  <span className="w-5 shrink-0 text-right text-xs text-subtle">{index + 1}</span>
                  <span className="min-w-0 flex-1 text-ink [overflow-wrap:anywhere]">{item.stock.name}</span>
                  <span className="shrink-0 font-semibold whitespace-nowrap text-ink">×{formatNumber(item.use_qty)}</span>
                </li>
              ))}
            </ol>
          )}
        </section>
      </div>

      <footer className="flex justify-end gap-3 border-t border-line px-5 py-4 sm:px-6">
        <button
          type="button"
          onClick={onClose}
          className="h-11 rounded-control border border-line-strong px-5 text-sm font-medium text-ink transition hover:bg-surface-2"
        >
          ปิด
        </button>
        {canManage && (
          <button
            type="button"
            onClick={onEdit}
            className="flex h-11 items-center gap-2 rounded-control bg-primary px-5 text-sm font-semibold text-white shadow-sm transition hover:bg-[#4d1621]"
          >
            <FiEdit2 className="size-4" />
            แก้ไขคอร์ส
          </button>
        )}
      </footer>
    </Modal>
  )
}

function Stat({ label, value }) {
  return (
    <div className="min-w-0 rounded-control bg-surface-2 px-3 py-2.5">
      <dt className="text-[11px] text-muted">{label}</dt>
      <dd className="mt-1 text-sm font-semibold text-ink [overflow-wrap:anywhere]">{value}</dd>
    </div>
  )
}

export default CourseDetailModal
