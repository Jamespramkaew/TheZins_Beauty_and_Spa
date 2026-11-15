import { FiEdit2 } from 'react-icons/fi'
import { formatBaht, formatNumber } from '../../../utils/format'
import { formatValidDay, pricePerSession, toNumber } from '../courseUtils'
import CategoryBadge from './CategoryBadge'
import CourseStatus from './CourseStatus'
import ItemChips from './ItemChips'

function CourseCard({ course, canManage, toggling, onEdit, onDelete, onToggleActive }) {
  return (
    <article
      className={`flex flex-col rounded-card border border-line p-5 transition hover:shadow-[0_6px_24px_-12px_rgba(23,24,28,0.18)] ${
        course.is_active ? 'bg-white' : 'bg-surface'
      }`}
    >
      <header className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <CategoryBadge category={course.category} />
          <h3 className="mt-2 text-[17px] font-semibold text-ink [overflow-wrap:anywhere]">{course.name}</h3>
        </div>
        {canManage && (
          <button
            type="button"
            onClick={onEdit}
            aria-label={`แก้ไขคอร์ส ${course.name}`}
            className="flex size-9 shrink-0 items-center justify-center rounded-control border border-line-strong text-muted transition hover:bg-surface-2 hover:text-ink"
          >
            <FiEdit2 className="size-4" />
          </button>
        )}
      </header>

      <p className="mt-4 flex flex-wrap items-baseline gap-x-2">
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

      <div className="mt-4 flex-1">
        <p className="mb-2 text-xs font-medium text-muted">สินค้าที่ใช้ต่อบริการ</p>
        <ItemChips items={course.items} />
      </div>

      <footer className="mt-4 flex items-center justify-between gap-3 border-t border-line pt-4">
        <CourseStatus course={course} canManage={canManage} pending={toggling} onToggle={onToggleActive} />
        {canManage && (
          <button type="button" onClick={onDelete} className="text-sm font-medium text-danger hover:underline">
            ลบ
          </button>
        )}
      </footer>
    </article>
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

export default CourseCard
