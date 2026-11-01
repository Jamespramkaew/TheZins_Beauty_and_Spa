import { LuCircleX, LuDroplet, LuPackage, LuSun, LuTriangleAlert } from 'react-icons/lu'
import Card from '../../../components/Card'
import CardLink from './CardLink'

const itemIcons = { sun: LuSun, droplet: LuDroplet }

function StockAlertsCard({ items }) {
  return (
    <Card
      title={
        <span className="flex items-center gap-2">
          สต็อกที่ต้องดู
          <span className="flex size-5 items-center justify-center rounded-full bg-danger-bg text-[11px] font-bold text-danger">
            {items.length}
          </span>
        </span>
      }
      aside="ทุกสาขา"
    >
      <ul className="space-y-4">
        {items.map((item) => {
          const outOfStock = item.remaining === 0
          const Icon = itemIcons[item.icon] ?? LuPackage

          return (
            <li key={item.id} className="flex items-center gap-3">
              <span
                className={`flex size-10 shrink-0 items-center justify-center rounded-control ${
                  outOfStock ? 'bg-danger-bg text-danger' : 'bg-warn-bg text-warn'
                }`}
              >
                <Icon className="size-4.5" />
              </span>
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium text-ink">{item.name}</p>
                <p className="text-xs text-muted">{item.note}</p>
              </div>
              {outOfStock ? (
                <span className="inline-flex items-center gap-1 rounded-md border border-danger/15 bg-danger-bg px-2 py-1 text-xs font-semibold text-danger">
                  <LuCircleX className="size-3.5" />
                  หมด
                </span>
              ) : (
                <span className="inline-flex items-center gap-1 rounded-md border border-warn/15 bg-warn-bg px-2 py-1 text-xs font-semibold text-warn">
                  <LuTriangleAlert className="size-3.5" />
                  เหลือ {item.remaining}
                </span>
              )}
            </li>
          )
        })}
      </ul>
      <CardLink to="/stocks">ไปที่ Stocks</CardLink>
    </Card>
  )
}

export default StockAlertsCard
