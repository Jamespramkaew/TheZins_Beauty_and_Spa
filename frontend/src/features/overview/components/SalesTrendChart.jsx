import {
  Area,
  CartesianGrid,
  ComposedChart,
  Line,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { LuArrowDown, LuArrowUp } from 'react-icons/lu'
import Card from '../../../components/Card'
import { formatBaht, formatNumber, formatPercent } from '../../../utils/format'
import { currentMonthLabel, previousMonthLabel } from '../mockData'

const CURRENT_COLOR = '#2F4C8F'
const PREVIOUS_COLOR = '#A3A6AF'
const monthShort = currentMonthLabel.split(' ')[0]

const axisTick = { fill: '#6B6D76', fontSize: 12 }

function SalesTrendChart({ data, total }) {
  const average = Math.round(total / data.length)
  const lastDay = data[data.length - 1].day
  const xTicks = [1, 8, 15, 22, lastDay]

  return (
    <Card className="lg:col-span-2">
      <div className="mb-2 flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-[15px] font-semibold text-ink">ยอดขาย 30 วัน</h2>
          <ul className="mt-2 flex items-center gap-4 text-xs text-muted">
            <li className="flex items-center gap-1.5">
              <span className="h-0.5 w-4 rounded-full" style={{ backgroundColor: CURRENT_COLOR }} />
              {currentMonthLabel}
            </li>
            <li className="flex items-center gap-1.5">
              <span className="w-4 border-t-2 border-dashed" style={{ borderColor: PREVIOUS_COLOR }} />
              {previousMonthLabel}
            </li>
          </ul>
        </div>
        <div className="text-right">
          <p className="text-2xl font-bold tracking-tight text-ink">{formatBaht(total)}</p>
          <p className="mt-1 text-xs text-muted">เฉลี่ย {formatBaht(average)} / วัน</p>
        </div>
      </div>

      <div className="-ml-2 h-70 min-w-0">
        <ResponsiveContainer width="100%" height="100%">
          <ComposedChart data={data} margin={{ top: 16, right: 12, bottom: 0, left: 0 }}>
            <defs>
              <linearGradient id="salesFill" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={CURRENT_COLOR} stopOpacity={0.16} />
                <stop offset="100%" stopColor={CURRENT_COLOR} stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid vertical={false} stroke="#EEEDF0" />
            <XAxis
              dataKey="day"
              ticks={xTicks}
              tickFormatter={(day) => `${day} ${monthShort}`}
              tick={axisTick}
              tickLine={false}
              axisLine={false}
              tickMargin={10}
            />
            <YAxis
              domain={[0, 45000]}
              ticks={[0, 15000, 30000, 45000]}
              tickFormatter={(value) => (value ? `${value / 1000}k` : '0')}
              tick={axisTick}
              tickLine={false}
              axisLine={false}
              width={44}
            />
            <Tooltip
              content={<TrendTooltip />}
              cursor={{ stroke: '#C9CBD3', strokeDasharray: '4 4' }}
            />
            <Line
              type="monotone"
              dataKey="previous"
              name={previousMonthLabel}
              stroke={PREVIOUS_COLOR}
              strokeWidth={1.5}
              strokeDasharray="5 4"
              dot={false}
              activeDot={{ r: 4, fill: PREVIOUS_COLOR, stroke: '#fff', strokeWidth: 2 }}
            />
            <Area
              type="monotone"
              dataKey="sales"
              name={currentMonthLabel}
              stroke={CURRENT_COLOR}
              strokeWidth={2}
              fill="url(#salesFill)"
              dot={false}
              activeDot={{ r: 5, fill: CURRENT_COLOR, stroke: '#fff', strokeWidth: 2 }}
            />
          </ComposedChart>
        </ResponsiveContainer>
      </div>
    </Card>
  )
}

function TrendTooltip({ active, payload }) {
  if (!active || !payload?.length) return null
  const { day, sales, previous } = payload[0].payload
  const change = ((sales - previous) / previous) * 100
  const up = change >= 0
  const ChangeIcon = up ? LuArrowUp : LuArrowDown

  return (
    <div className="min-w-48 rounded-control border border-line bg-white p-3 text-xs shadow-[0_12px_32px_-12px_rgba(23,24,28,0.25)]">
      <p className="mb-2 font-semibold text-ink">
        {day} {currentMonthLabel}
      </p>
      <TooltipRow color={CURRENT_COLOR} label="ยอดขาย" value={sales} />
      <TooltipRow color={PREVIOUS_COLOR} label="เดือนก่อน" value={previous} />
      <p className={`mt-2 flex items-center gap-1 border-t border-line pt-2 ${up ? 'text-chart' : 'text-danger'}`}>
        <ChangeIcon className="size-3" />
        <span className="font-semibold">
          {up ? '+' : '−'}
          {formatPercent(Math.abs(change))}
        </span>
        <span className="text-muted">จากเดือนก่อน</span>
      </p>
    </div>
  )
}

function TooltipRow({ color, label, value }) {
  return (
    <p className="flex items-center justify-between gap-6 py-0.5">
      <span className="flex items-center gap-1.5 text-muted">
        <span className="size-2 rounded-sm" style={{ backgroundColor: color }} />
        {label}
      </span>
      <span className="font-semibold text-ink tabular-nums">{formatNumber(value)}</span>
    </p>
  )
}

export default SalesTrendChart
