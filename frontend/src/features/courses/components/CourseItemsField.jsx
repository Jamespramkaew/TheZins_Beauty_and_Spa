import { useMemo, useState } from 'react'
import { FiAlertCircle, FiMinus, FiPlus, FiX } from 'react-icons/fi'
import { isContainer, maxUseQty } from '../courseUtils'
import StockPicker from './StockPicker'

// `rows` = useFieldArray fields (for keys), `values` = current items ({ stock_id, stock_name, use_qty })
function CourseItemsField({ rows, values, stocks, errors, changed, onAdd, onRemove, onQtyChange }) {
  const [picking, setPicking] = useState(false)
  const stockById = useMemo(() => new Map(stocks.list.map((stock) => [stock.id, stock])), [stocks.list])

  const selectedIds = new Set(values.map((item) => item.stock_id))
  // Already-linked stock may be inactive/deleted and still stays; only new picks must be active
  const available = stocks.list.filter((stock) => stock.is_active && !selectedIds.has(stock.id))

  const add = (stock) => {
    onAdd(stock)
    setPicking(false)
  }

  return (
    <section aria-labelledby="course-items-title" className="rounded-card border border-line bg-surface p-4 sm:p-5">
      <div className="flex flex-wrap items-center gap-2">
        <h3 id="course-items-title" className="font-semibold text-ink">
          สินค้าที่ใช้ต่อบริการ
        </h3>
        <span className="rounded-md bg-gold-bg px-2 py-0.5 text-[11px] font-semibold text-gold-ink">
          {values.length} รายการ
        </span>
        {changed && (
          <span className="rounded-md bg-[#EEF1F8] px-2 py-0.5 text-[11px] font-semibold text-chart">มีการเปลี่ยนแปลง</span>
        )}
      </div>
      <p className="mt-1 text-xs text-muted">จำนวนที่ระบบตัดสต็อกทุกครั้งที่ลูกค้าใช้บริการ 1 ครั้ง</p>

      <div className="mt-4 space-y-2.5">
        {values.length === 0 && (
          <div className="rounded-control border border-dashed border-line-strong bg-white px-4 py-6 text-center">
            <p className="text-sm font-semibold text-ink">ยังไม่ได้กำหนดสินค้าที่ใช้</p>
            <p className="mt-1 text-xs text-muted">คอร์สที่ไม่ใช้สินค้าเลยก็บันทึกได้</p>
          </div>
        )}

        {rows.length > 0 && (
          <ul className="space-y-2.5">
            {rows.map((row, index) => {
              const item = values[index]
              if (!item) return null
              return (
                <ItemRow
                  key={row.id}
                  item={item}
                  stock={stockById.get(item.stock_id)}
                  error={errors?.[index]?.stock_id?.message}
                  onQtyChange={(qty) => onQtyChange(index, qty)}
                  onRemove={() => onRemove(index)}
                />
              )
            })}
          </ul>
        )}

        {picking ? (
          <StockPicker stocks={available} status={stocks.status} onSelect={add} onClose={() => setPicking(false)} />
        ) : (
          <button
            type="button"
            onClick={() => setPicking(true)}
            className="flex h-11 w-full items-center justify-center gap-2 rounded-control border border-dashed border-line-strong bg-white text-sm font-medium text-ink transition hover:border-primary/40 hover:bg-surface-2"
          >
            <FiPlus className="size-4" />
            เพิ่มสินค้าจากสต็อก
          </button>
        )}
      </div>

      <p className="mt-3 text-xs leading-relaxed text-muted">
        เพิ่มได้เฉพาะสินค้าที่เปิดใช้งาน และเลือกซ้ำไม่ได้ · จำนวนขั้นต่ำ 1 · ของกระปุกใส่ได้ไม่เกินจำนวนครั้งที่ 1 กระปุกใช้ได้
      </p>
    </section>
  )
}

function ItemRow({ item, stock, error, onQtyChange, onRemove }) {
  const name = item.stock_name ?? stock?.name ?? 'สินค้า'
  const container = isContainer(stock)
  const max = maxUseQty(stock)
  const qty = Number(item.use_qty) || 1
  const clamp = (value) => Math.min(max, Math.max(1, value))

  return (
    <li className={`rounded-control border ${error ? 'border-danger/40 bg-danger-bg/40' : 'border-line bg-white'}`}>
      <div className="flex flex-wrap items-center gap-3 p-3.5 sm:flex-nowrap">
        <div className="min-w-0 flex-1">
          <p className="text-sm font-semibold text-ink [overflow-wrap:anywhere]">{name}</p>
          <span
            className={`mt-1.5 inline-flex rounded-md px-2 py-0.5 text-[11px] font-medium ${
              container ? 'bg-[#F3EDF7] text-[#5B3A7A]' : 'bg-[#EEF1F8] text-chart'
            }`}
          >
            {container ? `ของกระปุก · 1 กระปุกใช้ได้ ${stock.uses_per_unit} ครั้ง` : 'ของชิ้นเดียว'}
          </span>
        </div>

        <div className="flex flex-col items-end gap-1">
          <div className="flex h-9 items-center rounded-control border border-line-strong bg-white">
            <button
              type="button"
              aria-label={`ลดจำนวน ${name}`}
              disabled={qty <= 1}
              onClick={() => onQtyChange(clamp(qty - 1))}
              className="flex h-full w-9 items-center justify-center text-muted transition hover:text-ink disabled:opacity-30"
            >
              <FiMinus className="size-3.5" />
            </button>
            <input
              type="number"
              inputMode="numeric"
              min={1}
              max={Number.isFinite(max) ? max : undefined}
              value={qty}
              aria-label={`จำนวน ${name} ต่อการทำ 1 ครั้ง`}
              aria-invalid={error ? 'true' : undefined}
              onChange={(event) => {
                const value = parseInt(event.target.value, 10)
                if (!Number.isNaN(value)) onQtyChange(clamp(value))
              }}
              className={`h-full w-12 [appearance:textfield] bg-transparent text-center text-sm font-semibold focus:outline-none [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none ${
                error ? 'text-danger' : 'text-ink'
              }`}
            />
            <button
              type="button"
              aria-label={`เพิ่มจำนวน ${name}`}
              disabled={qty >= max}
              onClick={() => onQtyChange(clamp(qty + 1))}
              className="flex h-full w-9 items-center justify-center text-muted transition hover:text-ink disabled:opacity-30"
            >
              <FiPlus className="size-3.5" />
            </button>
          </div>
          <p className="text-[11px] text-muted">
            {container ? `ครั้ง / ครั้ง (1 กระปุกใช้ได้ ${stock.uses_per_unit} ครั้ง)` : 'ชิ้น / ครั้ง'}
          </p>
        </div>

        <button
          type="button"
          onClick={onRemove}
          aria-label="ลบแถวสินค้านี้"
          className="flex size-8 shrink-0 items-center justify-center rounded-lg text-danger transition hover:bg-danger-bg"
        >
          <FiX className="size-4" />
        </button>
      </div>

      {error && (
        <p className="mx-3.5 flex items-start gap-1.5 border-t border-dashed border-danger/30 py-2.5 text-xs text-danger">
          <FiAlertCircle className="mt-0.5 size-3.5 shrink-0" />
          {error}
        </p>
      )}
    </li>
  )
}

export default CourseItemsField
