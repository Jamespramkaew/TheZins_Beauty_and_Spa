import { FiMinus, FiPlus, FiX } from 'react-icons/fi'
import { formatBaht, formatNumber } from '../../../utils/format'
import { MONEY_PATTERN, PAYMENT_METHODS, SOURCES } from '../orderUtils'

// Right side of the sell screen: who it's for, the lines, totals, payment now, confirm.
// All values live in SellPage; this only shows them and reports changes.
function BillPanel({
  customer,
  onChangeCustomer,
  lines,
  lineErrors,
  onRemove,
  onQty,
  onDiscount,
  totals,
  payment,
  onPayment,
  orderDate,
  onOrderDate,
  today,
  source,
  onSource,
  error,
  submitting,
  onSubmit,
}) {
  const amountInvalid = payment.amount.trim() !== '' && !payment.valid

  return (
    <section aria-label="บิล" className="flex flex-col rounded-card border border-line bg-white xl:sticky xl:top-6">
      <div className="flex items-center justify-between gap-3 border-b border-line px-5 py-4">
        <div className="min-w-0">
          <p className="flex flex-wrap items-center gap-2 font-semibold text-ink [overflow-wrap:anywhere]">
            {[customer.fname, customer.lname].filter(Boolean).join(' ')}
            {customer.tier === 'vip' && (
              <span className="rounded-md bg-gold-bg px-2 py-0.5 text-xs font-medium text-gold-ink">VIP</span>
            )}
          </p>
          <p className="mt-0.5 text-xs text-muted">{customer.hn_no}</p>
        </div>
        <button type="button" onClick={onChangeCustomer} className="shrink-0 text-sm text-chart underline-offset-4 hover:underline">
          เปลี่ยน
        </button>
      </div>

      <div className="space-y-3 px-5 py-4">
        <div className="flex items-center justify-between">
          <h2 className="text-[15px] font-semibold text-ink">รายการในบิล</h2>
          <span className="text-xs text-muted">{formatNumber(lines.length)} รายการ</span>
        </div>

        {lines.length === 0 ? (
          <p className="rounded-control border border-dashed border-line-strong bg-surface px-4 py-8 text-center text-sm text-muted">
            ยังไม่มีรายการ — แตะการ์ดคอร์สหรือสินค้าทางซ้ายเพื่อเพิ่มลงบิล
          </p>
        ) : (
          <ul className="space-y-3">
            {lines.map((line, index) => {
              const gross = line.gross
              const discountInvalid =
                line.discount.trim() !== '' && (!MONEY_PATTERN.test(line.discount.trim()) || Number(line.discount) > gross)
              return (
                <li key={line.key} className="rounded-control border border-line p-3">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <p className="font-medium text-ink [overflow-wrap:anywhere]">{line.name}</p>
                      <p className="mt-0.5 text-xs text-muted">
                        {line.type === 'course'
                          ? `คอร์ส · ${formatNumber(line.sessions)} ครั้ง`
                          : `สินค้า · ${formatBaht(line.price)} / ชิ้น`}
                      </p>
                    </div>
                    <button
                      type="button"
                      onClick={() => onRemove(line.key)}
                      aria-label={`นำ ${line.name} ออกจากบิล`}
                      className="flex size-8 shrink-0 items-center justify-center rounded-control text-muted transition hover:bg-surface-2 hover:text-ink"
                    >
                      <FiX className="size-4" />
                    </button>
                  </div>

                  <div className="mt-3 flex flex-wrap items-end justify-between gap-3">
                    {line.type === 'stock' ? (
                      <div className="flex items-center rounded-control border border-line-strong">
                        <button
                          type="button"
                          onClick={() => onQty(line.key, line.qty - 1)}
                          disabled={line.qty <= 1}
                          aria-label={`ลดจำนวน ${line.name}`}
                          className="flex size-8 items-center justify-center text-muted disabled:opacity-40"
                        >
                          <FiMinus className="size-3.5" />
                        </button>
                        <span className="min-w-8 text-center text-sm font-medium">{formatNumber(line.qty)}</span>
                        <button
                          type="button"
                          onClick={() => onQty(line.key, line.qty + 1)}
                          aria-label={`เพิ่มจำนวน ${line.name}`}
                          className="flex size-8 items-center justify-center text-muted"
                        >
                          <FiPlus className="size-3.5" />
                        </button>
                      </div>
                    ) : (
                      <span />
                    )}
                    <label className="flex items-center gap-2 text-xs text-muted">
                      ส่วนลด ฿
                      <input
                        type="text"
                        inputMode="decimal"
                        value={line.discount}
                        onChange={(event) => onDiscount(line.key, event.target.value)}
                        placeholder="0"
                        aria-invalid={discountInvalid ? 'true' : undefined}
                        className={`h-8 w-24 rounded-control border px-2 text-right text-sm text-ink focus:ring-3 focus:outline-none ${
                          discountInvalid ? 'border-danger bg-danger-bg/30 focus:ring-danger/10' : 'border-line-strong focus:ring-chart/10'
                        }`}
                      />
                    </label>
                    <span className="text-sm font-semibold text-ink">{formatBaht(line.total)}</span>
                  </div>
                  {(discountInvalid || lineErrors[index]) && (
                    <p className="mt-2 text-xs text-danger">
                      {lineErrors[index] ?? `ส่วนลดต้องเป็นตัวเลข ทศนิยมไม่เกิน 2 ตำแหน่ง และไม่เกิน ${formatBaht(gross)}`}
                    </p>
                  )}
                </li>
              )
            })}
          </ul>
        )}
      </div>

      <dl className="space-y-1.5 border-t border-line px-5 py-4 text-sm">
        <div className="flex justify-between text-muted">
          <dt>ยอดรวม</dt>
          <dd>{formatBaht(totals.gross)}</dd>
        </div>
        <div className="flex justify-between text-muted">
          <dt>ส่วนลด</dt>
          <dd>{totals.discount > 0 ? `−${formatBaht(totals.discount)}` : formatBaht(0)}</dd>
        </div>
        <div className="flex items-baseline justify-between pt-1">
          <dt className="font-semibold text-ink">ยอดสุทธิ</dt>
          <dd className="text-[22px] font-bold text-ink">{formatBaht(totals.net)}</dd>
        </div>
      </dl>

      <div className="space-y-3 border-t border-line px-5 py-4">
        <p className="text-sm font-semibold text-ink">รับชำระตอนนี้</p>
        <div role="group" aria-label="วิธีชำระเงิน" className="grid grid-cols-3 gap-2">
          {PAYMENT_METHODS.map((method) => (
            <button
              key={method.value}
              type="button"
              aria-pressed={payment.method === method.value}
              onClick={() => onPayment({ method: method.value })}
              className={`h-10 rounded-control border text-sm transition ${
                payment.method === method.value
                  ? 'border-primary bg-primary/5 font-semibold text-ink'
                  : 'border-line-strong text-muted hover:text-ink'
              }`}
            >
              {method.label}
            </button>
          ))}
        </div>
        <label className="block text-xs text-muted">
          จำนวนเงิน ฿ (เว้นว่าง = ยังไม่ชำระ)
          <input
            type="text"
            inputMode="decimal"
            value={payment.amount}
            onChange={(event) => onPayment({ amount: event.target.value, touched: true })}
            aria-invalid={amountInvalid ? 'true' : undefined}
            className={`mt-1 h-10 w-full rounded-control border px-3 text-right text-sm text-ink focus:ring-3 focus:outline-none ${
              amountInvalid ? 'border-danger bg-danger-bg/30 focus:ring-danger/10' : 'border-line-strong focus:ring-chart/10'
            }`}
          />
        </label>
        {amountInvalid ? (
          <p className="text-xs text-danger">จำนวนเงินต้องมากกว่า 0 ทศนิยมไม่เกิน 2 ตำแหน่ง และไม่เกินยอดสุทธิ</p>
        ) : (
          <p className="text-xs text-muted">ค้างชำระหลังขาย {formatBaht(payment.remaining)}</p>
        )}

        <div className="grid grid-cols-2 gap-2 pt-1">
          <label className="text-xs text-muted">
            วันที่ขาย
            <input
              type="date"
              value={orderDate}
              max={today}
              onChange={(event) => onOrderDate(event.target.value || today)}
              className="mt-1 h-10 w-full rounded-control border border-line-strong px-2 text-sm text-ink focus:ring-3 focus:ring-chart/10 focus:outline-none"
            />
          </label>
          <label className="text-xs text-muted">
            ช่องทาง
            <select
              value={source}
              onChange={(event) => onSource(event.target.value)}
              className="mt-1 h-10 w-full rounded-control border border-line-strong bg-white px-2 text-sm text-ink focus:ring-3 focus:ring-chart/10 focus:outline-none"
            >
              {SOURCES.map((option) => (
                <option key={option.value || 'none'} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
        </div>
      </div>

      <div className="border-t border-line px-5 py-4">
        {error && (
          <p role="alert" className="mb-3 rounded-control bg-danger-bg px-3 py-2 text-sm text-danger">
            {error}
          </p>
        )}
        <button
          type="button"
          onClick={onSubmit}
          disabled={lines.length === 0 || submitting}
          className="h-12 w-full rounded-control bg-primary text-sm font-semibold text-white shadow-sm transition hover:bg-[#4d1621] disabled:cursor-not-allowed disabled:bg-line-strong disabled:text-subtle disabled:shadow-none"
        >
          {submitting ? 'กำลังบันทึก…' : lines.length === 0 ? 'ยืนยันการขาย' : `ยืนยันการขาย · ${formatBaht(totals.net)}`}
        </button>
      </div>
    </section>
  )
}

export default BillPanel
