import { useId, useState } from 'react'
import { useForm, useWatch } from 'react-hook-form'
import { FiAlertCircle, FiArrowRight } from 'react-icons/fi'
import Modal from '../../../components/Modal'
import { stockLogApi } from '../../../api/stockLogApi'
import { useToast } from '../../../hooks/useToast'
import { formatNumber } from '../../../utils/format'
import { mapStockError } from '../stockErrors'
import { branchColor } from '../stockUtils'
import { productOptions, useBranchQuantities, useEmployees } from '../useBranchData'
import ProductCombobox from './ProductCombobox'
import { inputClass, outlineButtonClass, primaryButtonClass } from './styles'
import { BranchDot, Field, ModalHeader, Stepper } from './ui'

// Backend field names for transfer differ from the movement form
const FIELD_ALIASES = { branch_id: 'from_branch_id' }

// preset: { stockId, branchId (origin) } from a table row (all optional)
function TransferModal({ isAdmin, branches, ownBranch, stocks, preset, onClose, onSaved }) {
  const toast = useToast()
  const formId = useId()
  const titleId = `${formId}-title`
  const [submitError, setSubmitError] = useState(false)

  const initialFrom = isAdmin ? (preset?.branchId ?? branches[0]?.id ?? '') : ownBranch.id
  const {
    register,
    control,
    handleSubmit,
    setValue,
    setError,
    clearErrors,
    formState: { errors, isSubmitting },
  } = useForm({
    defaultValues: {
      stock_id: preset?.stockId ?? '',
      from_branch_id: initialFrom,
      to_branch_id: branches.find((branch) => branch.id !== initialFrom)?.id ?? '',
      employee_id: '',
      qty: 1,
    },
  })
  const [stockId, fromId, toId, employeeId, qty] = useWatch({
    control,
    name: ['stock_id', 'from_branch_id', 'to_branch_id', 'employee_id', 'qty'],
  })

  const { employees, status: employeeStatus } = useEmployees(fromId, isAdmin)
  const origin = useBranchQuantities(fromId, isAdmin)
  // Branch accounts can't read other branches' stock, so only admin sees the destination side
  const destination = useBranchQuantities(toId, isAdmin, { enabled: isAdmin })
  const options = productOptions(stocks, origin.byStock, preset?.stockId)

  const branchName = (id) => branches.find((branch) => branch.id === id)?.name ?? (id === ownBranch.id ? ownBranch.name : '')
  const destinations = branches.filter((branch) => branch.id !== fromId)
  const amount = Number(qty) || 0
  const fromQty = origin.byStock.get(stockId)?.quantity
  const toQty = destination.byStock.get(stockId)?.quantity
  const exceeds = fromQty != null && amount > fromQty

  const blocker = !stockId
    ? 'กรุณาเลือกสินค้า'
    : !toId
      ? 'กรุณาเลือกสาขาปลายทาง'
      : !employeeId
        ? 'กรุณาเลือกพนักงานของสาขาต้นทาง'
        : amount < 1
          ? 'จำนวนต้องอย่างน้อย 1 ชิ้น'
          : exceeds
            ? 'จำนวนที่โอนมากกว่ายอดคงเหลือที่ต้นทาง'
            : null

  const onSubmit = async (values) => {
    setSubmitError(false)
    try {
      await stockLogApi.transfer({
        stock_id: values.stock_id,
        to_branch_id: values.to_branch_id,
        employee_id: values.employee_id,
        qty: Number(values.qty),
        ...(isAdmin && { from_branch_id: values.from_branch_id }),
      })
      toast.success(
        'โอนสำเร็จ',
        `โอน ${formatNumber(Number(values.qty))} ชิ้น จาก ${branchName(fromId)} ไป ${branchName(values.to_branch_id)} แล้ว`,
      )
      onSaved()
    } catch (error) {
      const mapped = mapStockError(error)
      mapped.forEach(({ field, message }) => setError(FIELD_ALIASES[field] ?? field, { type: 'server', message }))
      if (mapped.length === 0) setSubmitError(true)
    }
  }

  return (
    <Modal onClose={onClose} labelledBy={titleId} fullScreenOnMobile className="sm:max-w-[600px]">
      <form onSubmit={handleSubmit(onSubmit)} noValidate className="flex min-h-0 flex-1 flex-col">
        <ModalHeader
          titleId={titleId}
          title="โอนสาขา"
          subtitle="ระบบบันทึกเป็น 2 รายการ คือ ออกจากต้นทาง และเข้าปลายทาง"
          onClose={onClose}
        />

        <div className="min-h-0 flex-1 space-y-5 overflow-y-auto px-5 py-5 sm:px-6">
          {submitError && (
            <p
              role="alert"
              className="flex items-start gap-3 rounded-control border border-danger/25 bg-danger-bg px-4 py-3 text-sm text-danger"
            >
              <FiAlertCircle className="mt-0.5 size-4 shrink-0" />
              โอนไม่สำเร็จ — ยังไม่มีข้อมูลใดถูกบันทึก กรุณาลองใหม่อีกครั้ง
            </p>
          )}

          <Field id={`${formId}-stock`} label="สินค้า" required error={errors.stock_id?.message}>
            <ProductCombobox
              id={`${formId}-stock`}
              value={stockId}
              options={options}
              status={origin.status}
              invalid={Boolean(errors.stock_id)}
              onChange={(id) => {
                setValue('stock_id', id)
                clearErrors(['stock_id', 'qty'])
              }}
            />
          </Field>

          <div className="grid items-center gap-3 sm:grid-cols-[1fr_auto_1fr]">
            <div className="rounded-card border border-line bg-surface p-3.5">
              <Field id={`${formId}-from`} label="จากสาขา" required error={errors.from_branch_id?.message}>
                {isAdmin ? (
                  <select
                    id={`${formId}-from`}
                    className={inputClass(errors.from_branch_id)}
                    {...register('from_branch_id', {
                      onChange: (event) => {
                        setValue('employee_id', '')
                        if (toId === event.target.value) {
                          setValue('to_branch_id', branches.find((branch) => branch.id !== event.target.value)?.id ?? '')
                        }
                        clearErrors(['from_branch_id', 'employee_id', 'qty'])
                      },
                    })}
                  >
                    {branches.map((branch) => (
                      <option key={branch.id} value={branch.id}>
                        {branch.name}
                      </option>
                    ))}
                  </select>
                ) : (
                  <select id={`${formId}-from`} disabled value={ownBranch.id} className={inputClass(false)}>
                    <option value={ownBranch.id}>{ownBranch.name}</option>
                  </select>
                )}
              </Field>
              <p className="mt-1.5 text-xs text-muted">
                {stockId && fromQty != null ? (
                  <>
                    คงเหลือที่ต้นทาง <span className="font-semibold text-ink">{formatNumber(fromQty)} ชิ้น</span>
                  </>
                ) : isAdmin ? (
                  'เลือกสินค้าเพื่อดูยอดคงเหลือ'
                ) : (
                  'บัญชีสาขาโอนออกได้เฉพาะสาขาตัวเอง'
                )}
              </p>
            </div>

            <FiArrowRight aria-hidden="true" className="mx-auto size-5 rotate-90 text-muted sm:rotate-0" />

            <div className="rounded-card border border-line bg-surface p-3.5">
              <Field id={`${formId}-to`} label="ไปสาขา" required error={errors.to_branch_id?.message}>
                <select
                  id={`${formId}-to`}
                  className={inputClass(errors.to_branch_id)}
                  {...register('to_branch_id', { onChange: () => clearErrors('to_branch_id') })}
                >
                  {destinations.length === 0 && <option value="">ไม่มีสาขาอื่น</option>}
                  {destinations.map((branch) => (
                    <option key={branch.id} value={branch.id}>
                      {branch.name}
                    </option>
                  ))}
                </select>
              </Field>
              <p className="mt-1.5 text-xs text-muted">ไม่แสดงสาขาต้นทาง</p>
            </div>
          </div>

          <div className="grid items-start gap-4 sm:grid-cols-2">
            <div>
              <label htmlFor={`${formId}-qty`} className="mb-1.5 block text-sm font-medium text-ink">
                จำนวน<span className="ml-0.5 text-danger">*</span>
              </label>
              <Stepper
                id={`${formId}-qty`}
                label="จำนวน"
                value={qty}
                invalid={exceeds || Boolean(errors.qty)}
                onChange={(value) => {
                  setValue('qty', value)
                  clearErrors('qty')
                }}
              />
              {(exceeds || errors.qty) && (
                <p className="mt-1.5 text-xs font-medium text-danger">
                  {errors.qty?.message ?? `คงเหลือที่ต้นทางแค่ ${formatNumber(fromQty)} ชิ้น`}
                </p>
              )}
            </div>
            <Field
              id={`${formId}-employee`}
              label="พนักงานของสาขาต้นทาง"
              required
              hint={employeeStatus === 'error' ? 'โหลดรายชื่อพนักงานไม่สำเร็จ' : undefined}
              error={errors.employee_id?.message}
            >
              <select
                id={`${formId}-employee`}
                disabled={employeeStatus !== 'ready'}
                className={inputClass(errors.employee_id)}
                {...register('employee_id', { required: 'กรุณาเลือกพนักงานของสาขาต้นทาง' })}
              >
                <option value="">
                  {employeeStatus === 'loading' ? 'กำลังโหลด…' : employees.length ? 'เลือกพนักงาน' : 'ไม่มีพนักงานในสาขานี้'}
                </option>
                {employees.map((employee) => (
                  <option key={employee.id} value={employee.id}>
                    {employee.name}
                  </option>
                ))}
              </select>
            </Field>
          </div>

          <div>
            <p className="mb-2 text-xs text-muted">ผลหลังโอน</p>
            <div className={`grid gap-2.5 ${isAdmin ? 'sm:grid-cols-2' : ''}`}>
              <PreviewBox
                color={branchColor(branches, fromId)}
                name={branchName(fromId)}
                before={stockId ? fromQty : null}
                after={fromQty == null ? null : fromQty - amount}
                tone="text-danger"
              />
              {isAdmin && (
                <PreviewBox
                  color={branchColor(branches, toId)}
                  name={branchName(toId)}
                  before={stockId ? toQty : null}
                  after={toQty == null ? null : toQty + amount}
                  tone="text-success"
                />
              )}
            </div>
            <p className="mt-2 text-xs text-muted">
              หมายเหตุระบบใส่ให้เอง: “Transfer from {branchName(fromId) || 'A'} to {branchName(toId) || 'B'}”
            </p>
          </div>
        </div>

        <footer className="flex flex-wrap items-center justify-end gap-3 border-t border-line px-5 py-4 sm:px-6">
          {blocker && <p className="mr-auto text-xs text-muted">{blocker}</p>}
          <button type="button" onClick={onClose} className={outlineButtonClass}>
            ยกเลิก
          </button>
          <button type="submit" disabled={Boolean(blocker) || isSubmitting} className={primaryButtonClass}>
            {isSubmitting ? 'กำลังโอน…' : `โอน ${formatNumber(amount)} ชิ้น`}
          </button>
        </footer>
      </form>
    </Modal>
  )
}

function PreviewBox({ color, name, before, after, tone }) {
  return (
    <div className="flex items-center gap-2 rounded-card bg-surface-2 px-4 py-3 text-sm tabular-nums">
      <BranchDot color={color} />
      <span className="font-semibold text-ink">{name}:</span>
      {before == null ? (
        <span className="text-subtle">—</span>
      ) : (
        <>
          <span className="text-muted">{formatNumber(before)}</span>
          <FiArrowRight className="size-3.5 text-muted" />
          <span className={`font-bold ${tone}`}>{formatNumber(after)}</span>
        </>
      )}
    </div>
  )
}

export default TransferModal
