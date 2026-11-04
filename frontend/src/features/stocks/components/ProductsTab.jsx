import { useState } from 'react'
import { FiEdit2, FiEye, FiInfo, FiPlus, FiSearch, FiTrash2 } from 'react-icons/fi'
import Toggle from '../../../components/ui/Toggle'
import { formatBaht, formatNumber } from '../../../utils/format'
import { toNumber } from '../stockUtils'
import { primaryButtonClass } from './styles'
import { CategoryText, EmptyState, InfoBox, LoadError, NoResults, SkeletonRows } from './ui'

// GET /stocks is not paginated, so search runs in the browser
function ProductsTab({ isAdmin, stocks, togglingIds, onCreate, onEdit, onDelete, onToggleActive }) {
  const [search, setSearch] = useState('')
  const keyword = search.trim().toLowerCase()
  const rows = (stocks.data ?? []).filter((stock) => !keyword || stock.name.toLowerCase().includes(keyword))

  return (
    <div className="space-y-5">
      <InfoBox icon={FiInfo}>
        หน้านี้คือข้อมูลสินค้า ยอดคงเหลือดูได้ที่แท็บ <span className="font-semibold text-ink">‘ยอดคงเหลือ’</span>
      </InfoBox>

      <div className="flex flex-wrap items-center justify-between gap-3">
        <label className="relative w-full max-w-sm">
          <FiSearch className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-subtle" />
          <input
            type="search"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="ค้นหาสินค้า"
            aria-label="ค้นหาสินค้า"
            className="h-10 w-full rounded-control border border-line-strong bg-white pr-3 pl-10 text-sm placeholder:text-subtle focus:border-chart/40 focus:ring-3 focus:ring-chart/10 focus:outline-none"
          />
        </label>
        {isAdmin ? (
          <button type="button" onClick={onCreate} className={primaryButtonClass}>
            <FiPlus className="size-4" />
            เพิ่มสินค้า
          </button>
        ) : (
          <span className="flex items-center gap-2 rounded-full border border-line bg-surface-2 px-3.5 py-1.5 text-sm text-muted">
            <FiEye className="size-4" />
            ดูอย่างเดียว (บัญชีสาขา)
          </span>
        )}
      </div>

      {stocks.error && !stocks.loaded ? (
        <LoadError error={stocks.error} onRetry={stocks.reload} />
      ) : !stocks.loaded ? (
        <SkeletonRows />
      ) : rows.length === 0 ? (
        keyword ? (
          <NoResults onClear={() => setSearch('')} />
        ) : (
          <EmptyState title="ยังไม่มีสินค้า" detail={isAdmin ? 'กด “เพิ่มสินค้า” เพื่อเริ่มต้น' : undefined} />
        )
      ) : (
        <div className="overflow-x-auto rounded-card border border-line">
          <table className="w-full min-w-[1080px] table-fixed text-sm">
            <colgroup>
              <col />
              <col className="w-[120px]" />
              <col className="w-[110px]" />
              <col className="w-[170px]" />
              <col className="w-[110px]" />
              <col className="w-[100px]" />
              <col className="w-[160px]" />
              {isAdmin && <col className="w-[130px]" />}
            </colgroup>
            <thead className="text-xs text-muted">
              <tr className="border-b border-line text-left whitespace-nowrap">
                <th className="py-3.5 pr-3.5 pl-5 font-medium">ชื่อสินค้า</th>
                <th className="px-3.5 py-3.5 font-medium">หมวด</th>
                <th className="px-3.5 py-3.5 text-right font-medium">ราคา</th>
                <th className="px-3.5 py-3.5 font-medium">การขาย</th>
                <th className="px-3.5 py-3.5 text-right font-medium">ใช้ได้ / ชิ้น</th>
                <th className="px-3.5 py-3.5 text-right font-medium">จุดเตือน</th>
                <th className={`py-3.5 pl-3.5 font-medium ${isAdmin ? 'pr-3.5' : 'pr-5'}`}>สถานะ</th>
                {isAdmin && <th className="py-3.5 pr-5 pl-3.5 text-right font-medium">การทำงาน</th>}
              </tr>
            </thead>
            <tbody>
              {rows.map((stock) => {
                const dim = stock.is_active ? '' : 'opacity-55'
                return (
                  <tr key={stock.id} className="border-b border-line bg-white transition-colors last:border-0 hover:bg-surface">
                    <td className={`py-3.5 pr-3.5 pl-5 font-semibold text-ink [overflow-wrap:anywhere] ${dim}`}>{stock.name}</td>
                    <td className={`px-3.5 py-3.5 ${dim}`}>
                      <CategoryText category={stock.category} />
                    </td>
                    <td className={`px-3.5 py-3.5 text-right font-semibold tabular-nums ${dim}`}>
                      {stock.is_sellable ? formatBaht(toNumber(stock.price)) : <span className="text-subtle">—</span>}
                    </td>
                    <td className={`px-3.5 py-3.5 ${dim} ${stock.is_sellable ? 'text-ink' : 'text-muted'}`}>
                      {stock.is_sellable ? 'ขายได้' : 'ใช้ในบริการเท่านั้น'}
                    </td>
                    <td className={`px-3.5 py-3.5 text-right tabular-nums ${dim}`}>{formatNumber(stock.uses_per_unit)} ครั้ง</td>
                    <td className={`px-3.5 py-3.5 text-right tabular-nums ${dim}`}>{formatNumber(stock.min_qty)} ชิ้น</td>
                    <td className={`py-3.5 pl-3.5 ${isAdmin ? 'pr-3.5' : 'pr-5'}`}>
                      {isAdmin ? (
                        <Toggle
                          checked={stock.is_active}
                          disabled={togglingIds.has(stock.id)}
                          onChange={() => onToggleActive(stock)}
                          label={`เปิดใช้งานสินค้า ${stock.name}`}
                        >
                          <span className={`whitespace-nowrap ${stock.is_active ? '' : 'text-muted'}`}>
                            {stock.is_active ? 'ใช้งาน' : 'ปิดใช้งาน'}
                          </span>
                        </Toggle>
                      ) : (
                        <span
                          className={`inline-flex rounded-full px-2.5 py-1 text-xs font-medium ${
                            stock.is_active ? 'bg-success-bg text-success' : 'bg-surface-2 text-muted'
                          }`}
                        >
                          {stock.is_active ? 'ใช้งาน' : 'ปิดใช้งาน'}
                        </span>
                      )}
                    </td>
                    {isAdmin && (
                      <td className="py-2.5 pr-5 pl-3.5">
                        <div className="flex justify-end gap-2">
                          <IconButton label={`แก้ไขสินค้า ${stock.name}`} onClick={() => onEdit(stock)}>
                            <FiEdit2 className="size-4" />
                          </IconButton>
                          <IconButton label={`ลบสินค้า ${stock.name}`} onClick={() => onDelete(stock)} danger>
                            <FiTrash2 className="size-4" />
                          </IconButton>
                        </div>
                      </td>
                    )}
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

function IconButton({ label, onClick, danger, children }) {
  return (
    <button
      type="button"
      aria-label={label}
      title={label}
      onClick={onClick}
      className={`flex size-10 items-center justify-center rounded-control border border-line-strong bg-white transition hover:bg-surface-2 focus-visible:ring-3 focus-visible:ring-chart/20 focus-visible:outline-none ${
        danger ? 'text-danger' : 'text-ink'
      }`}
    >
      {children}
    </button>
  )
}

export default ProductsTab
