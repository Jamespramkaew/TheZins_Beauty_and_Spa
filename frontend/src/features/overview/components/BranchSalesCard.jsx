import { Link } from 'react-router-dom'
import { LuArrowRight } from 'react-icons/lu'
import Card from '../../../components/Card'
import { useAuth } from '../../../hooks/useAuth'
import { formatBaht, formatPercent } from '../../../utils/format'

function BranchSalesCard({ branches }) {
  const { user } = useAuth()
  const total = branches.reduce((sum, branch) => sum + branch.sales, 0)
  const max = Math.max(...branches.map((branch) => branch.sales))

  return (
    <Card title="ยอดขายตามสาขา" aside="เดือนนี้">
      {/* Share of total; segments are separated by a 2px gap */}
      <div className="flex h-2 gap-0.5" role="img" aria-label="สัดส่วนยอดขายแต่ละสาขา">
        {branches.map((branch) => (
          <span
            key={branch.id}
            className="rounded-full"
            style={{ width: `${(branch.sales / total) * 100}%`, backgroundColor: branch.color }}
          />
        ))}
      </div>

      <ul className="mt-5 flex-1 space-y-5">
        {branches.map((branch) => (
          <li key={branch.id}>
            <div className="flex items-center gap-2.5 text-sm">
              <span className="size-2.5 rounded-sm" style={{ backgroundColor: branch.color }} />
              <span className="flex-1 font-medium text-ink">{branch.name}</span>
              <span className="text-xs text-muted tabular-nums">
                {formatPercent((branch.sales / total) * 100)}
              </span>
              <span className="w-24 text-right font-semibold text-ink tabular-nums">
                {formatBaht(branch.sales)}
              </span>
            </div>
            <div className="mt-2 h-1 rounded-full bg-[#F0F0F3]">
              <div
                className="h-full rounded-full"
                style={{ width: `${(branch.sales / max) * 100}%`, backgroundColor: branch.color }}
              />
            </div>
          </li>
        ))}
      </ul>

      {user.role === 'admin' && (
        <Link
          to="/branch-accounts"
          className="mt-6 flex items-center justify-between rounded-control bg-surface-2 px-4 py-3 text-sm text-ink transition hover:bg-line"
        >
          ดูบัญชีแต่ละสาขา
          <LuArrowRight className="size-4" />
        </Link>
      )}
    </Card>
  )
}

export default BranchSalesCard
