import { LuStar } from 'react-icons/lu'
import Card from '../../../components/Card'
import { formatBaht, formatPercent } from '../../../utils/format'
import CardLink from './CardLink'

function TopEmployeeCard({ employee, branch, monthTotal }) {
  return (
    <Card title="พนักงานขายสูงสุด" aside="เดือนนี้">
      <div className="flex items-center gap-3">
        <span className="relative flex size-12 shrink-0 items-center justify-center rounded-full border border-gold/50 bg-gold-bg text-base font-bold text-gold-ink">
          {employee.initials}
          <span className="absolute -right-0.5 -bottom-0.5 flex size-4.5 items-center justify-center rounded-full bg-gold text-white ring-2 ring-white">
            <LuStar className="size-2.5 fill-current" />
          </span>
        </span>
        <div className="min-w-0">
          <p className="truncate text-sm font-semibold text-ink">{employee.name}</p>
          <p className="mt-0.5 flex items-center gap-1.5 text-xs text-muted">
            <span className="size-2 rounded-full" style={{ backgroundColor: branch.color }} />
            สาขา{branch.name}
          </p>
        </div>
      </div>

      <dl className="mt-5 grid grid-cols-2 gap-3">
        <div className="rounded-control bg-surface-2 p-3">
          <dt className="text-xs text-muted">ยอดขาย</dt>
          <dd className="mt-1 text-lg font-bold text-ink tabular-nums">{formatBaht(employee.sales)}</dd>
        </div>
        <div className="rounded-control bg-surface-2 p-3">
          <dt className="text-xs text-muted">สัดส่วนของยอดรวม</dt>
          <dd className="mt-1 text-lg font-bold text-ink tabular-nums">
            {formatPercent((employee.sales / monthTotal) * 100)}
          </dd>
        </div>
      </dl>
      <CardLink to="/employees">ดูพนักงานทั้งหมด</CardLink>
    </Card>
  )
}

export default TopEmployeeCard
