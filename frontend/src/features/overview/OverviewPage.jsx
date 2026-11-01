import { useState } from 'react'
import { currentUser } from '../../mocks/session'
import BranchFilter from './components/BranchFilter'
import BranchSalesCard from './components/BranchSalesCard'
import KpiCards from './components/KpiCards'
import SalesTrendChart from './components/SalesTrendChart'
import StockAlertsCard from './components/StockAlertsCard'
import TopCoursesCard from './components/TopCoursesCard'
import TopEmployeeCard from './components/TopEmployeeCard'
import {
  branches,
  dailySales,
  kpis,
  periodOptions,
  stockAlerts,
  topCourses,
  topEmployee,
} from './mockData'

const sum = (values) => values.reduce((total, value) => total + value, 0)

const monthTotal = sum(dailySales.map((day) => day.sales))
const previousMonthTotal = sum(dailySales.map((day) => day.previous))
const last7Days = dailySales.slice(-7)

const summary = {
  today: dailySales[dailySales.length - 1].sales,
  weekTotal: sum(last7Days.map((day) => day.sales)),
  last7Days,
  monthTotal,
  monthGrowth: ((monthTotal - previousMonthTotal) / previousMonthTotal) * 100,
}

function OverviewPage() {
  const [period, setPeriod] = useState('month')
  const [selectedBranchIds, setSelectedBranchIds] = useState(() => branches.map((branch) => branch.id))

  const selectedBranches = branches.filter((branch) => selectedBranchIds.includes(branch.id))
  const allBranchesSelected = selectedBranches.length === branches.length
  const employeeBranch = branches.find((branch) => branch.id === topEmployee.branchId)

  return (
    <div className="mx-auto max-w-[1400px] space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-[28px] font-bold tracking-tight text-ink">ภาพรวม</h1>
          <p className="mt-1 text-sm text-muted">
            สวัสดีค่ะ {currentUser.firstName} — สรุปยอดของ
            {allBranchesSelected ? `ทั้ง ${branches.length} สาขา` : `สาขา${selectedBranches.map((b) => b.name).join(', ')}`}
            ประจำเดือนกันยายน
          </p>
        </div>

        <div role="tablist" aria-label="ช่วงเวลา" className="flex rounded-[12px] border border-line bg-surface-2 p-1">
          {periodOptions.map((option) => (
            <button
              key={option.value}
              type="button"
              role="tab"
              aria-selected={period === option.value}
              onClick={() => setPeriod(option.value)}
              className={`h-9 rounded-lg px-3.5 text-sm transition ${
                period === option.value
                  ? 'bg-white font-semibold text-ink shadow-sm'
                  : 'text-muted hover:text-ink'
              }`}
            >
              {option.label}
            </button>
          ))}
        </div>
      </div>

      <BranchFilter branches={branches} selectedIds={selectedBranchIds} onChange={setSelectedBranchIds} />

      <KpiCards
        summary={summary}
        kpis={kpis}
        selectedBranches={selectedBranches}
        allBranchesSelected={allBranchesSelected}
      />

      <div className="grid gap-4 lg:grid-cols-3">
        <SalesTrendChart data={dailySales} total={summary.monthTotal} />
        <BranchSalesCard branches={selectedBranches} />
      </div>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <TopCoursesCard courses={topCourses} />
        <TopEmployeeCard employee={topEmployee} branch={employeeBranch} monthTotal={summary.monthTotal} />
        <StockAlertsCard items={stockAlerts} />
      </div>
    </div>
  )
}

export default OverviewPage
