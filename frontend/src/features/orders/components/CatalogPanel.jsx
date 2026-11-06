import { useState } from 'react'
import { FiCheck, FiPlus, FiSearch } from 'react-icons/fi'
import FilterGroup from '../../../components/ui/FilterGroup'
import { formatBaht, formatNumber } from '../../../utils/format'
import { toNumber } from '../../../utils/number'

const TABS = [
  { value: 'course', label: 'คอร์ส' },
  { value: 'stock', label: 'สินค้า' },
]

// Left side of the sell screen: pick courses (one of each) and products (qty adds up)
function CatalogPanel({ courses, stocks, lines, onToggleCourse, onAddStock }) {
  const [tab, setTab] = useState('course')
  const [search, setSearch] = useState('')
  const keyword = search.trim().toLowerCase()
  const match = (item) => !keyword || item.name.toLowerCase().includes(keyword)

  const inBill = (type, id) => lines.find((line) => line.type === type && line.id === id)
  const shownCourses = courses.filter(match)
  const shownStocks = stocks.filter(match)

  return (
    <section className="space-y-4">
      <div className="flex flex-wrap items-center gap-3">
        <label className="relative w-full max-w-sm">
          <FiSearch className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-subtle" />
          <input
            type="search"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="ค้นหาคอร์สหรือสินค้า"
            aria-label="ค้นหาคอร์สหรือสินค้า"
            className="h-10 w-full rounded-control border border-line-strong bg-white pr-3 pl-10 text-sm placeholder:text-subtle focus:border-chart/40 focus:ring-3 focus:ring-chart/10 focus:outline-none"
          />
        </label>
        <FilterGroup
          label="ประเภทรายการ"
          options={TABS}
          value={tab}
          counts={{ course: courses.length, stock: stocks.length }}
          onChange={setTab}
        />
      </div>

      {tab === 'course' ? (
        shownCourses.length === 0 ? (
          <Empty text={keyword ? 'ไม่พบคอร์สที่ค้นหา' : 'ยังไม่มีคอร์สที่เปิดขาย'} />
        ) : (
          <div className="grid gap-3 sm:grid-cols-2">
            {shownCourses.map((course) => {
              const selected = Boolean(inBill('course', course.id))
              const price = toNumber(course.price)
              return (
                <button
                  key={course.id}
                  type="button"
                  aria-pressed={selected}
                  onClick={() => onToggleCourse(course)}
                  className={`flex flex-col rounded-card border bg-white p-4 text-left transition ${
                    selected ? 'border-primary bg-primary/[0.03] ring-3 ring-primary/10' : 'border-line hover:border-line-strong'
                  }`}
                >
                  <span className="flex items-start justify-between gap-3">
                    <span className="font-semibold text-ink [overflow-wrap:anywhere]">{course.name}</span>
                    <span
                      className={`flex size-6 shrink-0 items-center justify-center rounded-full ${
                        selected ? 'bg-primary text-white' : 'border border-line-strong text-muted'
                      }`}
                    >
                      {selected ? <FiCheck className="size-3.5" /> : <FiPlus className="size-3.5" />}
                    </span>
                  </span>
                  <span className="mt-3 flex items-end justify-between gap-3 text-sm">
                    <span className="text-muted">
                      {formatNumber(course.max_session)} ครั้ง · {formatBaht(Math.round(price / course.max_session))} / ครั้ง
                    </span>
                    <span className="text-base font-bold text-ink">{formatBaht(price)}</span>
                  </span>
                </button>
              )
            })}
          </div>
        )
      ) : shownStocks.length === 0 ? (
        <Empty text={keyword ? 'ไม่พบสินค้าที่ค้นหา' : 'ยังไม่มีสินค้าที่เปิดขาย'} />
      ) : (
        <div className="grid gap-3 sm:grid-cols-2">
          {shownStocks.map((stock) => {
            const line = inBill('stock', stock.id)
            return (
              <button
                key={stock.id}
                type="button"
                onClick={() => onAddStock(stock)}
                className={`flex items-center justify-between gap-3 rounded-card border bg-white p-4 text-left transition ${
                  line ? 'border-primary bg-primary/[0.03]' : 'border-line hover:border-line-strong'
                }`}
              >
                <span className="min-w-0">
                  <span className="block font-semibold text-ink [overflow-wrap:anywhere]">{stock.name}</span>
                  <span className="mt-1 block text-sm text-muted">
                    {formatBaht(toNumber(stock.price))} / ชิ้น
                    {line && ` · ในบิล ${formatNumber(line.qty)} ชิ้น`}
                  </span>
                </span>
                <span className="flex size-8 shrink-0 items-center justify-center rounded-full border border-line-strong text-muted">
                  <FiPlus className="size-4" />
                </span>
              </button>
            )
          })}
        </div>
      )}
    </section>
  )
}

function Empty({ text }) {
  return (
    <p className="rounded-card border border-dashed border-line-strong bg-surface px-6 py-12 text-center text-sm text-muted">{text}</p>
  )
}

export default CatalogPanel
