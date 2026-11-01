import { Bar, BarChart, Cell, ResponsiveContainer, Tooltip } from 'recharts'
import { LuArrowDown, LuArrowUp, LuClock, LuTrendingUp, LuUserPlus, LuWallet } from 'react-icons/lu'
import { formatBaht, formatNumber, formatPercent } from '../../../utils/format'
import { currentMonthLabel } from '../mockData'

function KpiCards({ summary, kpis, selectedBranches, allBranchesSelected }) {
  const growthUp = summary.monthGrowth >= 0
  const GrowthIcon = growthUp ? LuArrowUp : LuArrowDown

  return (
    <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <StatCard label="ยอดขายวันนี้" icon={LuWallet} value={formatNumber(summary.today)} unit="฿"
        aside={<WeeklyBars data={summary.last7Days} />}
      >
        <span>
          สัปดาห์นี้ <span className="font-semibold text-ink">{formatBaht(summary.weekTotal)}</span>
        </span>
      </StatCard>

      <StatCard label="ยอดขายเดือนนี้" icon={LuTrendingUp} value={formatNumber(summary.monthTotal)} unit="฿">
        <span
          className={`mr-2 inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-xs font-semibold ${
            growthUp ? 'bg-[#EEF1F8] text-chart' : 'bg-danger-bg text-danger'
          }`}
        >
          <GrowthIcon className="size-3" />
          {formatPercent(Math.abs(summary.monthGrowth), 0)}
        </span>
        จากเดือนก่อน
      </StatCard>

      <StatCard label="คิววันนี้" icon={LuClock} value={kpis.todayQueues} unit="คิว">
        <BranchScope branches={selectedBranches} all={allBranchesSelected} />
      </StatCard>

      <StatCard label="ลูกค้าใหม่เดือนนี้" icon={LuUserPlus} value={kpis.newCustomers} unit="คน" iconTone="gold">
        <BranchScope branches={selectedBranches} all={allBranchesSelected} />
      </StatCard>
    </div>
  )
}

function StatCard({ label, icon: Icon, value, unit, aside, iconTone = 'chart', children }) {
  return (
    <article className="rounded-card border border-line bg-white p-5 transition hover:shadow-[0_6px_24px_-12px_rgba(23,24,28,0.18)]">
      <div className="flex items-center justify-between">
        <h3 className="text-sm text-muted">{label}</h3>
        <span
          className={`flex size-8 items-center justify-center rounded-lg ${
            iconTone === 'gold' ? 'bg-gold-bg text-gold-ink' : 'bg-[#EEF1F8] text-chart'
          }`}
        >
          <Icon className="size-4" />
        </span>
      </div>
      <div className="mt-4 flex items-end justify-between gap-3">
        <p className="text-[28px] leading-none font-bold tracking-tight text-ink">
          {value}
          <span className="ml-1 text-base font-semibold">{unit}</span>
        </p>
        {aside}
      </div>
      <p className="mt-4 flex items-center text-xs text-muted">{children}</p>
    </article>
  )
}

function WeeklyBars({ data }) {
  const lastIndex = data.length - 1

  return (
    <div className="h-8 w-16" aria-label="ยอดขาย 7 วันล่าสุด">
      <ResponsiveContainer width="100%" height="100%">
        <BarChart data={data} barCategoryGap={2} margin={{ top: 0, right: 0, bottom: 0, left: 0 }}>
          <Tooltip
            cursor={false}
            allowEscapeViewBox={{ x: true, y: true }}
            wrapperStyle={{ zIndex: 10 }}
            content={<WeeklyBarsTooltip />}
          />
          <Bar dataKey="sales" radius={[2, 2, 0, 0]} isAnimationActive={false}>
            {data.map((entry, index) => (
              <Cell key={entry.day} fill={index === lastIndex ? '#2F4C8F' : '#C9D2E6'} />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </div>
  )
}

function WeeklyBarsTooltip({ active, payload }) {
  if (!active || !payload?.length) return null
  const { day, sales } = payload[0].payload

  return (
    <div className="rounded-lg border border-line bg-white px-2.5 py-1.5 text-xs whitespace-nowrap shadow-lg">
      <span className="text-muted">{day} {currentMonthLabel.split(' ')[0]}</span>{' '}
      <span className="font-semibold text-ink">{formatBaht(sales)}</span>
    </div>
  )
}

function BranchScope({ branches, all }) {
  return (
    <>
      <span className="mr-2 flex -space-x-0.5">
        {branches.map((branch) => (
          <span key={branch.id} className="size-2 rounded-full ring-1 ring-white" style={{ backgroundColor: branch.color }} />
        ))}
      </span>
      {all ? 'ทุกสาขา' : branches.map((branch) => branch.name).join(', ')}
    </>
  )
}

export default KpiCards
