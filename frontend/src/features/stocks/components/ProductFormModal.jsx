import { useId, useState } from 'react'
import { useForm, useWatch } from 'react-hook-form'
import { FiAlertCircle } from 'react-icons/fi'
import Modal from '../../../components/Modal'
import Toggle from '../../../components/ui/Toggle'
import { stockApi } from '../../../api/stockApi'
import { useToast } from '../../../hooks/useToast'
import { mapStockError } from '../stockErrors'
import { toNumber } from '../stockUtils'
import { inputClass, outlineButtonClass, primaryButtonClass } from './styles'
import { Field, ModalHeader, SegmentedControl } from './ui'

const NAME_MAX = 150

const CATEGORY_OPTIONS = [
  { value: 'premium', label: 'Premium' },
  { value: 'normal', label: 'Normal' },
]

const isWhole = (value) => /^\d+$/.test(String(value).trim())

const toFormValues = (stock) =>
  stock
    ? {
        name: stock.name,
        category: stock.category,
        is_sellable: stock.is_sellable,
        price: String(toNumber(stock.price)),
        uses_per_unit: String(stock.uses_per_unit),
        min_qty: String(stock.min_qty),
        is_active: stock.is_active,
      }
    : { name: '', category: 'normal', is_sellable: true, price: '', uses_per_unit: '1', min_qty: '0', is_active: true }

// stock = null → create
function ProductFormModal({ stock, onClose, onSaved }) {
  const isEdit = Boolean(stock)
  const toast = useToast()
  const formId = useId()
  const titleId = `${formId}-title`
  const [submitFailed, setSubmitFailed] = useState(false)

  const {
    register,
    control,
    handleSubmit,
    setValue,
    setError,
    clearErrors,
    formState: { errors, isSubmitting },
  } = useForm({ defaultValues: toFormValues(stock) })
  const [name, category, isSellable, isActive] = useWatch({
    control,
    name: ['name', 'category', 'is_sellable', 'is_active'],
  })

  const onSubmit = async (values) => {
    setSubmitFailed(false)
    const body = {
      stock_name: values.name.trim(),
      category: values.category,
      is_sellable: values.is_sellable,
      // Not sellable = service-only; the backend forces price to 0 anyway
      price: values.is_sellable ? String(values.price).trim() : '0',
      uses_per_unit: Number(values.uses_per_unit),
      min_qty: Number(values.min_qty || 0),
    }

    try {
      const response = isEdit
        ? await stockApi.update(stock.id, { ...body, is_active: values.is_active })
        : await stockApi.create(body)
      toast.success('บันทึกสำเร็จ', `${isEdit ? 'แก้ไข' : 'เพิ่ม'}สินค้า “${response.data.name}” เรียบร้อยแล้ว`)
      onSaved(response.data)
    } catch (error) {
      // Nothing was saved — keep the modal open with what the user typed
      mapStockError(error).forEach(({ field, message }) => setError(field, { type: 'server', message }))
      setSubmitFailed(true)
    }
  }

  const nameLength = name.trim().length

  return (
    <Modal onClose={onClose} labelledBy={titleId} fullScreenOnMobile className="sm:max-w-[560px]">
      <form onSubmit={handleSubmit(onSubmit)} noValidate className="flex min-h-0 flex-1 flex-col">
        <ModalHeader titleId={titleId} title={isEdit ? 'แก้ไขสินค้า' : 'เพิ่มสินค้า'} onClose={onClose} />

        <div className="min-h-0 flex-1 space-y-5 overflow-y-auto px-5 py-5 sm:px-6">
          {submitFailed && !errors.name && (
            <p
              role="alert"
              className="flex items-start gap-3 rounded-control border border-danger/25 bg-danger-bg px-4 py-3 text-sm text-danger"
            >
              <FiAlertCircle className="mt-0.5 size-4 shrink-0" />
              บันทึกไม่สำเร็จ — ยังไม่มีข้อมูลใดถูกบันทึก ตรวจสอบข้อมูลแล้วลองอีกครั้ง
            </p>
          )}

          <div>
            <div className="mb-1.5 flex items-baseline justify-between">
              <label htmlFor={`${formId}-name`} className="text-sm font-medium text-ink">
                ชื่อสินค้า<span className="ml-0.5 text-danger">*</span>
              </label>
              <span className={`text-xs tabular-nums ${nameLength > NAME_MAX ? 'text-danger' : 'text-subtle'}`}>
                {nameLength}/{NAME_MAX}
              </span>
            </div>
            <input
              id={`${formId}-name`}
              type="text"
              placeholder="เช่น เซรั่มวิตามินซี"
              aria-invalid={errors.name ? 'true' : undefined}
              className={inputClass(errors.name)}
              {...register('name', {
                validate: (value) => {
                  const trimmed = value.trim()
                  if (!trimmed) return 'กรุณากรอกชื่อสินค้า'
                  if (trimmed.length > NAME_MAX) return `ชื่อสินค้ายาวได้ไม่เกิน ${NAME_MAX} ตัวอักษร`
                  return true
                },
              })}
            />
            {errors.name && (
              <p className="mt-1.5 flex items-center gap-1.5 rounded-control border border-danger/25 bg-danger-bg px-3 py-2 text-xs text-danger">
                <FiAlertCircle className="size-3.5 shrink-0" />
                {errors.name.message}
              </p>
            )}
          </div>

          <div>
            <p className="mb-1.5 text-sm font-medium text-ink">
              หมวด<span className="ml-0.5 text-danger">*</span>
            </p>
            <div className="inline-flex">
              <SegmentedControl
                label="หมวด"
                options={CATEGORY_OPTIONS}
                value={category}
                onChange={(value) => setValue('category', value, { shouldDirty: true })}
              />
            </div>
          </div>

          <SwitchRow
            title="ขายหน้าร้านได้"
            detail="ปิดไว้ถ้าเป็นของที่ใช้ในบริการอย่างเดียว"
            checked={isSellable}
            onChange={(checked) => {
              setValue('is_sellable', checked, { shouldDirty: true })
              if (!checked) setValue('price', '0')
              clearErrors('price')
            }}
          />

          <div className="grid gap-4 sm:grid-cols-3">
            <Field
              id={`${formId}-price`}
              label="ราคา (฿)"
              required={isSellable}
              hint={isSellable ? undefined : 'สินค้าใช้ในบริการเท่านั้น ราคาจะเป็น 0'}
              error={errors.price?.message}
            >
              <input
                id={`${formId}-price`}
                type="number"
                inputMode="decimal"
                min={0}
                step="0.01"
                disabled={!isSellable}
                className={inputClass(errors.price)}
                {...register('price', {
                  validate: (value, values) => {
                    if (!values.is_sellable) return true
                    if (String(value).trim() === '') return 'กรุณากรอกราคา'
                    return Number(value) >= 0 || 'ราคาต้องไม่ติดลบ'
                  },
                })}
              />
            </Field>
            <Field
              id={`${formId}-uses`}
              label="ใช้ได้กี่ครั้งต่อชิ้น"
              required
              hint="มากกว่า 1 = ของกระปุก เช่น เซรั่ม 1 ขวดใช้ได้ 10 ครั้ง"
              error={errors.uses_per_unit?.message}
            >
              <input
                id={`${formId}-uses`}
                type="number"
                inputMode="numeric"
                min={1}
                className={inputClass(errors.uses_per_unit)}
                {...register('uses_per_unit', {
                  validate: (value) => (isWhole(value) && Number(value) >= 1) || 'ต้องเป็นจำนวนเต็มอย่างน้อย 1',
                })}
              />
            </Field>
            <Field
              id={`${formId}-min`}
              label="จุดเตือนสั่งของ"
              hint="เตือนเมื่อชิ้นที่ยังไม่เปิดเหลือ ≤ ค่านี้"
              error={errors.min_qty?.message}
            >
              <input
                id={`${formId}-min`}
                type="number"
                inputMode="numeric"
                min={0}
                className={inputClass(errors.min_qty)}
                {...register('min_qty', {
                  validate: (value) => String(value).trim() === '' || isWhole(value) || 'ต้องเป็นจำนวนเต็ม 0 ขึ้นไป',
                })}
              />
            </Field>
          </div>

          {isEdit && (
            <SwitchRow
              title="เปิดใช้งาน"
              detail="ปิดแล้วจะไม่แสดงตอนเลือกสินค้า แต่ประวัติเดิมยังอยู่ครบ"
              checked={isActive}
              onChange={(checked) => setValue('is_active', checked, { shouldDirty: true })}
            />
          )}
        </div>

        <footer className="flex justify-end gap-3 border-t border-line px-5 py-4 sm:px-6">
          <button type="button" onClick={onClose} className={outlineButtonClass}>
            ยกเลิก
          </button>
          <button type="submit" disabled={isSubmitting} className={`${primaryButtonClass} disabled:cursor-wait`}>
            {isSubmitting ? 'กำลังบันทึก…' : isEdit ? 'บันทึกการแก้ไข' : 'เพิ่มสินค้า'}
          </button>
        </footer>
      </form>
    </Modal>
  )
}

function SwitchRow({ title, detail, checked, onChange }) {
  return (
    <div className="flex items-center justify-between gap-4 rounded-card border border-line p-4">
      <div>
        <p className="text-sm font-medium text-ink">{title}</p>
        <p className="mt-0.5 text-xs text-muted">{detail}</p>
      </div>
      <Toggle checked={checked} onChange={onChange} label={title} />
    </div>
  )
}

export default ProductFormModal
