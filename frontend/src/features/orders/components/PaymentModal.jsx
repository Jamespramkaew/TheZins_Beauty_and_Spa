import { useId, useState } from 'react'
import { useForm, useWatch } from 'react-hook-form'
import { FiAlertCircle, FiX } from 'react-icons/fi'
import Modal from '../../../components/Modal'
import { orderApi } from '../../../api/orderApi'
import { useToast } from '../../../hooks/useToast'
import { formatBaht } from '../../../utils/format'
import { toNumber } from '../../../utils/number'
import { mapPaymentError } from '../orderErrors'
import { PAYMENT_METHODS } from '../orderUtils'

// Same rule as the backend: more than 0, at most 2 decimals
const MONEY_PATTERN = /^\d+(\.\d{1,2})?$/

// Records one more payment. The backend re-checks the outstanding amount under a row lock,
// so two cashiers paying at once can't overpay the bill.
function PaymentModal({ order, onClose, onPaid }) {
  const toast = useToast()
  const formId = useId()
  const titleId = `${formId}-title`
  const [submitFailed, setSubmitFailed] = useState(false)
  const outstanding = toNumber(order.outstanding)

  const {
    register,
    control,
    handleSubmit,
    setError,
    setValue,
    formState: { errors, isSubmitting },
  } = useForm({ defaultValues: { method: 'cash', amount: String(outstanding), reference: '' } })
  const amount = useWatch({ control, name: 'amount' })
  const remaining = Math.max(0, outstanding - (MONEY_PATTERN.test(String(amount).trim()) ? Number(amount) : 0))

  const onSubmit = async (values) => {
    setSubmitFailed(false)
    try {
      const response = await orderApi.addPayment(order.id, {
        method: values.method,
        amount: values.amount.trim(),
        reference: values.reference.trim(),
      })
      toast.success('รับชำระแล้ว', `${formatBaht(Number(values.amount))} บิล ${order.order_no}`)
      onPaid(response.data)
    } catch (error) {
      // Nothing was saved — keep the modal open with what the user typed
      mapPaymentError(error).forEach(({ field, message }) => setError(field, { type: 'server', message }))
      setSubmitFailed(true)
    }
  }

  return (
    <Modal onClose={onClose} labelledBy={titleId} className="max-w-[520px]">
      <form onSubmit={handleSubmit(onSubmit)} noValidate className="flex min-h-0 flex-1 flex-col">
        <header className="flex items-start justify-between gap-4 border-b border-line px-6 py-4">
          <div>
            <h2 id={titleId} className="text-lg font-bold text-ink">
              รับชำระเพิ่ม
            </h2>
            <p className="mt-0.5 text-[13px] text-muted">
              บิล {order.order_no} · ค้างชำระ <span className="font-semibold text-danger">{formatBaht(outstanding)}</span>
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="ปิด"
            className="flex size-9 shrink-0 items-center justify-center rounded-control bg-surface-2 text-muted transition hover:text-ink"
          >
            <FiX className="size-4" />
          </button>
        </header>

        <div className="min-h-0 flex-1 space-y-5 overflow-y-auto px-6 py-5">
          {submitFailed && (
            <div
              role="alert"
              className="flex items-start gap-3 rounded-control border border-danger/25 bg-danger-bg px-4 py-3 text-sm text-danger"
            >
              <FiAlertCircle className="mt-0.5 size-4 shrink-0" />
              <p>
                <span className="font-semibold">บันทึกไม่สำเร็จ — ยังไม่มีการรับชำระ</span>
                <br />
                {Object.keys(errors).length > 0 ? 'แก้ไขจุดที่มีข้อความสีแดง แล้วกดบันทึกอีกครั้ง' : 'กรุณาลองใหม่อีกครั้ง'}
              </p>
            </div>
          )}

          <fieldset>
            <legend className="mb-1.5 text-sm font-medium text-ink">
              วิธีชำระ<span className="ml-0.5 text-danger">*</span>
            </legend>
            <div className="grid grid-cols-3 gap-2">
              {PAYMENT_METHODS.map((method) => (
                <label
                  key={method.value}
                  className="flex h-11 cursor-pointer items-center justify-center rounded-control border border-line-strong text-sm text-ink transition has-checked:border-primary has-checked:bg-primary/5 has-checked:font-semibold has-focus-visible:ring-3 has-focus-visible:ring-primary/20"
                >
                  <input type="radio" value={method.value} className="sr-only" {...register('method')} />
                  {method.label}
                </label>
              ))}
            </div>
            {errors.method && <p className="mt-1.5 text-xs text-danger">{errors.method.message}</p>}
          </fieldset>

          <div>
            <label htmlFor={`${formId}-amount`} className="mb-1.5 block text-sm font-medium text-ink">
              จำนวนเงิน (฿)<span className="ml-0.5 text-danger">*</span>
            </label>
            <div className="flex gap-2">
              <input
                id={`${formId}-amount`}
                type="text"
                inputMode="decimal"
                aria-invalid={errors.amount ? 'true' : undefined}
                className={inputClass(errors.amount)}
                {...register('amount', {
                  validate: (value) => {
                    const text = value.trim()
                    if (!MONEY_PATTERN.test(text) || Number(text) <= 0) {
                      return 'จำนวนเงินต้องมากกว่า 0 และมีทศนิยมไม่เกิน 2 ตำแหน่ง'
                    }
                    if (Number(text) > outstanding) return `รับได้ไม่เกินยอดค้าง ${formatBaht(outstanding)}`
                    return true
                  },
                })}
              />
              <button
                type="button"
                onClick={() => setValue('amount', String(outstanding), { shouldValidate: true })}
                className="h-11 shrink-0 rounded-control border border-line-strong px-3 text-sm text-ink transition hover:bg-surface-2"
              >
                เต็มยอดค้าง
              </button>
            </div>
            {errors.amount ? (
              <p className="mt-1.5 text-xs text-danger">{errors.amount.message}</p>
            ) : (
              <p className="mt-1.5 text-xs text-muted">หลังรับชำระ ค้างอีก {formatBaht(remaining)}</p>
            )}
          </div>

          <div>
            <label htmlFor={`${formId}-reference`} className="mb-1.5 block text-sm font-medium text-ink">
              เลขอ้างอิง <span className="font-normal text-muted">(ไม่บังคับ)</span>
            </label>
            <input
              id={`${formId}-reference`}
              type="text"
              placeholder="เช่น เลขสลิป หรือเลขท้ายบัตร"
              aria-invalid={errors.reference ? 'true' : undefined}
              className={inputClass(errors.reference)}
              {...register('reference', {
                validate: (value) => value.trim().length <= 100 || 'เลขอ้างอิงยาวได้ไม่เกิน 100 ตัวอักษร',
              })}
            />
            {errors.reference && <p className="mt-1.5 text-xs text-danger">{errors.reference.message}</p>}
          </div>
        </div>

        <footer className="flex justify-end gap-3 border-t border-line px-6 py-4">
          <button
            type="button"
            onClick={onClose}
            className="h-11 rounded-control border border-line-strong px-5 text-sm font-medium text-ink transition hover:bg-surface-2"
          >
            ยกเลิก
          </button>
          <button
            type="submit"
            disabled={isSubmitting}
            className="h-11 rounded-control bg-primary px-5 text-sm font-semibold text-white shadow-sm transition hover:bg-[#4d1621] disabled:cursor-wait disabled:opacity-70"
          >
            {isSubmitting ? 'กำลังบันทึก…' : 'บันทึกการรับชำระ'}
          </button>
        </footer>
      </form>
    </Modal>
  )
}

const inputClass = (error) =>
  `h-11 w-full rounded-control border px-3.5 text-sm text-ink placeholder:text-subtle focus:ring-3 focus:outline-none ${
    error
      ? 'border-danger bg-danger-bg/30 focus:ring-danger/10'
      : 'border-line-strong bg-white focus:border-chart/40 focus:ring-chart/10'
  }`

export default PaymentModal
