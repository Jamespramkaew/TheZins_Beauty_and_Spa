import { useId, useState } from 'react'
import { useForm, useWatch } from 'react-hook-form'
import { FiAlertCircle, FiArrowRight } from 'react-icons/fi'
import Modal from '../../../components/Modal'
import { stockLogApi } from '../../../api/stockLogApi'
import { useToast } from '../../../hooks/useToast'
import { formatNumber } from '../../../utils/format'
import { mapStockError } from '../stockErrors'
import { MOVEMENT_TYPES } from '../stockUtils'
import { productOptions, useBranchQuantities, useEmployees } from '../useBranchData'
import ProductCombobox from './ProductCombobox'
import { inputClass, outlineButtonClass, primaryButtonClass } from './styles'
import { Field, ModalHeader, SegmentedControl, Stepper } from './ui'

const ADJUST_OPTIONS = [
  { value: '+', label: '+ เพิ่มยอด' },
  { value: '-', label: '− ลดยอด' },
]

// preset: { type, stockId, branchId } from a table row (all optional)
function MovementModal({ isAdmin, branches, ownBranch, stocks, preset, onClose, onSaved }) {
  const toast = useToast()
  const formId = useId()
  const titleId = `${formId}-title`
  const [submitError, setSubmitError] = useState(false)

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
      type: preset?.type ?? 'receive',
      branch_id: isAdmin ? (preset?.branchId ?? branches[0]?.id ?? '') : ownBranch.id,
      employee_id: '',
      stock_id: preset?.stockId ?? '',
      adjust: '+',
      qty: 1,
      note: '',
    },
  })
  const [type, branchId, employeeId, stockId, adjust, qty, note] = useWatch({
    control,
    name: ['type', 'branch_id', 'employee_id', 'stock_id', 'adjust', 'qty', 'note'],
  })

  const { employees, status: employeeStatus } = useEmployees(branchId, isAdmin)
  const { byStock, status: qtyStatus } = useBranchQuantities(branchId, isAdmin)
  const options = productOptions(stocks, byStock, preset?.stockId)

  const movement = MOVEMENT_TYPES.find((item) => item.value === type)
  const sign = movement.sign === 0 ? (adjust === '-' ? -1 : 1) : movement.sign
  const amount = Number(qty) || 0
  const available = byStock.get(stockId)?.quantity
  const branchName = branches.find((branch) => branch.id === branchId)?.name ?? ownBranch.name
  const exceeds = sign < 0 && available != null && amount > available
  const noteRequired = type === 'adjust'

  // First reason the form can't be sent yet — shown beside the buttons
  const blocker = !employeeId
    ? 'กรุณาเลือกพนักงานที่ทำรายการ'
    : !stockId
      ? 'กรุณาเลือกสินค้า'
      : amount < 1
        ? 'จำนวนต้องอย่างน้อย 1 ชิ้น'
        : exceeds
          ? 'จำนวนที่ออกมากกว่ายอดคงเหลือ'
          : noteRequired && !note.trim()
            ? 'กรุณาระบุหมายเหตุของการปรับยอด'
            : null

  const onSubmit = async (values) => {
    setSubmitError(false)
    const body = {
      type: values.type,
      stock_id: values.stock_id,
      employee_id: values.employee_id,
      // Fixed-sign types: server applies the sign and requires qty > 0. Only adjust sends a signed qty.
      qty: movement.sign === 0 ? sign * Number(values.qty) : Number(values.qty),
      note: values.note.trim(),
      ...(isAdmin && { branch_id: values.branch_id }),
    }
    try {
      await stockLogApi.addMovement(body)
      const verb = values.type === 'adjust' ? `ปรับยอด ${sign > 0 ? '+' : '−'}` : `บันทึก${movement.label} `
      toast.success('บันทึกสำเร็จ', `${verb}${formatNumber(Number(values.qty))} ชิ้นแล้ว`)
      onSaved()
    } catch (error) {
      const mapped = mapStockError(error)
      mapped.forEach(({ field, message }) => setError(field, { type: 'server', message }))
      if (mapped.length === 0) setSubmitError(true)
    }
  }

  return (
    <Modal onClose={onClose} labelledBy={titleId} fullScreenOnMobile className="sm:max-w-[640px]">
      <form onSubmit={handleSubmit(onSubmit)} noValidate className="flex min-h-0 flex-1 flex-col">
        <ModalHeader
          titleId={titleId}
          title="บันทึกรายการ"
          subtitle="ขาย ใช้ในบริการ เปิดกระปุก และผลต่างนับสต็อก ระบบบันทึกให้เอง"
          onClose={onClose}
        />

        <div className="min-h-0 flex-1 space-y-5 overflow-y-auto px-5 py-5 sm:px-6">
          {submitError && (
            <p
              role="alert"
              className="flex items-start gap-3 rounded-control border border-danger/25 bg-danger-bg px-4 py-3 text-sm text-danger"
            >
              <FiAlertCircle className="mt-0.5 size-4 shrink-0" />
              บันทึกไม่สำเร็จ — ยังไม่มีข้อมูลใดถูกบันทึก กรุณาลองใหม่อีกครั้ง
            </p>
          )}

          <fieldset>
            <legend className="mb-2 text-sm font-medium text-ink">
              ประเภทรายการ<span className="ml-0.5 text-danger">*</span>
            </legend>
            <div role="radiogroup" className="grid grid-cols-2 gap-2.5 sm:grid-cols-5">
              {MOVEMENT_TYPES.map((item) => {
                const selected = item.value === type
                return (
                  <button
                    key={item.value}
                    type="button"
                    role="radio"
                    aria-checked={selected}
                    onClick={() => {
                      setValue('type', item.value)
                      clearErrors(['qty', 'note'])
                    }}
                    className={`flex flex-col items-start gap-1 rounded-control border p-3 text-left transition focus-visible:ring-3 focus-visible:ring-primary/20 focus-visible:outline-none ${
                      selected ? 'border-primary bg-[#FBF7F8]' : 'border-line-strong bg-white hover:bg-surface'
                    }`}
                  >
                    <item.icon className={`size-4 ${selected ? 'text-primary' : 'text-subtle'}`} />
                    <span className="mt-1 text-sm font-semibold text-ink">{item.label}</span>
                    <span className="text-[11px] leading-snug text-muted">{item.hint}</span>
                  </button>
                )
              })}
            </div>
          </fieldset>

          <div className="grid gap-4 sm:grid-cols-2">
            <Field
              id={`${formId}-branch`}
              label="สาขา"
              required
              hint={isAdmin ? undefined : 'บัญชีสาขาบันทึกได้เฉพาะสาขาตัวเอง'}
              error={errors.branch_id?.message}
            >
              {isAdmin ? (
                <select
                  id={`${formId}-branch`}
                  className={inputClass(errors.branch_id)}
                  {...register('branch_id', {
                    onChange: () => {
                      setValue('employee_id', '')
                      clearErrors(['branch_id', 'employee_id', 'qty'])
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
                // Locked: the backend uses the session's branch regardless of what is sent
                <select id={`${formId}-branch`} disabled value={ownBranch.id} className={inputClass(false)}>
                  <option value={ownBranch.id}>{ownBranch.name}</option>
                </select>
              )}
            </Field>
            <Field
              id={`${formId}-employee`}
              label="พนักงานที่ทำรายการ"
              required
              hint={employeeStatus === 'error' ? 'โหลดรายชื่อพนักงานไม่สำเร็จ' : 'เฉพาะพนักงานของสาขาที่เลือก'}
              error={errors.employee_id?.message}
            >
              <select
                id={`${formId}-employee`}
                disabled={employeeStatus !== 'ready'}
                className={inputClass(errors.employee_id)}
                {...register('employee_id', { required: 'กรุณาเลือกพนักงานที่ทำรายการ' })}
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

          <Field id={`${formId}-stock`} label="สินค้า" required error={errors.stock_id?.message}>
            <ProductCombobox
              id={`${formId}-stock`}
              value={stockId}
              options={options}
              status={qtyStatus}
              invalid={Boolean(errors.stock_id)}
              onChange={(id) => {
                setValue('stock_id', id)
                clearErrors(['stock_id', 'qty'])
              }}
            />
          </Field>

          {type === 'adjust' && (
            <div>
              <p className="mb-1.5 text-sm font-medium text-ink">
                ทิศทางการปรับ<span className="ml-0.5 text-danger">*</span>
              </p>
              <div className="inline-flex">
                <SegmentedControl
                  label="ทิศทางการปรับ"
                  options={ADJUST_OPTIONS}
                  value={adjust}
                  onChange={(value) => {
                    setValue('adjust', value)
                    clearErrors('qty')
                  }}
                />
              </div>
            </div>
          )}

          <div className="grid items-start gap-4 sm:grid-cols-[auto_1fr]">
            <div>
              <label htmlFor={`${formId}-qty`} className="mb-1.5 block text-sm font-medium text-ink">
                จำนวนชิ้น<span className="ml-0.5 text-danger">*</span>
              </label>
              <Stepper
                id={`${formId}-qty`}
                label="จำนวนชิ้น"
                value={qty}
                invalid={exceeds || Boolean(errors.qty)}
                onChange={(value) => {
                  setValue('qty', value)
                  clearErrors('qty')
                }}
              />
              {(exceeds || errors.qty) && (
                <p className="mt-1.5 text-xs font-medium text-danger">
                  {errors.qty?.message ?? `คงเหลือแค่ ${formatNumber(available)} ชิ้น`}
                </p>
              )}
            </div>

            <Preview
              branchName={branchName}
              available={stockId ? available : null}
              change={sign * amount}
              danger={exceeds}
            />
          </div>

          <Field
            id={`${formId}-note`}
            label="หมายเหตุ"
            required={noteRequired}
            optional={!noteRequired}
            error={errors.note?.message}
          >
            <textarea
              id={`${formId}-note`}
              rows={3}
              placeholder={noteRequired ? 'เหตุผลที่ต้องปรับยอด เช่น บันทึกรับเข้าซ้ำ' : 'เช่น เลขล็อต หรือรายละเอียดเพิ่มเติม'}
              className={`${inputClass(errors.note)} h-auto py-2.5`}
              {...register('note', {
                validate: (value, values) => values.type !== 'adjust' || value.trim() !== '' || 'กรุณาระบุเหตุผลของการปรับยอด',
              })}
            />
          </Field>
        </div>

        <footer className="flex flex-wrap items-center justify-end gap-3 border-t border-line px-5 py-4 sm:px-6">
          {blocker && <p className="mr-auto text-xs text-muted">{blocker}</p>}
          <button type="button" onClick={onClose} className={outlineButtonClass}>
            ยกเลิก
          </button>
          <button type="submit" disabled={Boolean(blocker) || isSubmitting} className={primaryButtonClass}>
            {isSubmitting ? 'กำลังบันทึก…' : `บันทึก${movement.label}`}
          </button>
        </footer>
      </form>
    </Modal>
  )
}

function Preview({ branchName, available, change, danger }) {
  const after = available == null ? null : available + change
  return (
    <div className={`rounded-card px-4 py-3.5 ${danger ? 'bg-danger-bg/60' : 'bg-surface-2'}`}>
      <p className="text-xs text-muted">ผลหลังบันทึก{branchName ? ` · ${branchName}` : ''}</p>
      {available == null ? (
        <p className="mt-1.5 text-sm text-subtle">เลือกสินค้าเพื่อดูยอดคงเหลือ</p>
      ) : (
        <p className="mt-1.5 flex flex-wrap items-center gap-2 text-[15px] text-muted tabular-nums">
          คงเหลือ {formatNumber(available)}
          <FiArrowRight className="size-4" />
          <span className={`text-lg font-bold ${change >= 0 ? 'text-success' : 'text-danger'}`}>{formatNumber(after)}</span>
          {change !== 0 && (
            <span
              className={`rounded-md px-1.5 py-0.5 text-xs font-semibold ${
                change > 0 ? 'bg-success-bg text-success' : 'bg-danger-bg text-danger'
              }`}
            >
              {change > 0 ? '+' : '−'}
              {formatNumber(Math.abs(change))}
            </span>
          )}
        </p>
      )}
    </div>
  )
}

export default MovementModal
