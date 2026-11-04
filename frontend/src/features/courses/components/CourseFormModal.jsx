import { useEffect, useId, useState } from 'react'
import { useFieldArray, useForm, useWatch } from 'react-hook-form'
import dayjs from 'dayjs'
import buddhistEra from 'dayjs/plugin/buddhistEra'
import 'dayjs/locale/th'
import { FiAlertCircle, FiInfo, FiX } from 'react-icons/fi'
import Modal from '../../../components/Modal'
import { courseApi } from '../../../api/courseApi'
import { stockApi } from '../../../api/stockApi'
import { useToast } from '../../../hooks/useToast'
import { formatBaht } from '../../../utils/format'
import { mapCourseError } from '../courseErrors'
import { pricePerSession, toNumber } from '../courseUtils'
import CourseItemsField from './CourseItemsField'
import Toggle from '../../../components/ui/Toggle'

dayjs.extend(buddhistEra)

const VALID_DAY_PRESETS = [
  { label: '90 วัน', days: 90 },
  { label: '180 วัน', days: 180 },
  { label: '1 ปี', days: 365 },
]

const toFormValues = (course) =>
  course
    ? {
        name: course.name,
        category: course.category ?? '',
        price: String(toNumber(course.price)),
        max_session: String(course.max_session),
        service_comm: String(toNumber(course.service_comm)),
        valid_day: course.valid_day == null ? '' : String(course.valid_day),
        no_expiry: course.valid_day == null,
        is_active: course.is_active,
        items: course.items.map((item) => ({
          stock_id: item.stock.id,
          stock_name: item.stock.name,
          use_qty: item.use_qty,
        })),
      }
    : {
        name: '',
        category: '',
        price: '',
        max_session: '1',
        service_comm: '0',
        valid_day: '',
        no_expiry: true,
        is_active: true,
        items: [],
      }

const isWhole = (value) => /^\d+$/.test(String(value).trim())

const countErrors = (errors) =>
  Object.entries(errors).reduce((count, [key, value]) => {
    if (key === 'root') return count
    if (key === 'items') return count + (Array.isArray(value) ? value.filter(Boolean).length : 1)
    return count + 1
  }, 0)

// course = null → create
function CourseFormModal({ course, categories, onClose, onSaved }) {
  const isEdit = Boolean(course)
  const toast = useToast()
  const formId = useId()
  const titleId = `${formId}-title`
  const [stocks, setStocks] = useState({ status: 'loading', list: [] })
  // Drives both the "มีการเปลี่ยนแปลง" badge and whether PATCH sends `items`
  const [itemsChanged, setItemsChanged] = useState(false)
  const [submitFailed, setSubmitFailed] = useState(false)

  const {
    register,
    control,
    handleSubmit,
    setValue,
    setError,
    clearErrors,
    formState: { errors, isSubmitting },
  } = useForm({ defaultValues: toFormValues(course) })
  const { fields, append, remove } = useFieldArray({ control, name: 'items' })
  const items = useWatch({ control, name: 'items' })
  const [price, maxSession, serviceComm, validDay, noExpiry, isActive] = useWatch({
    control,
    name: ['price', 'max_session', 'service_comm', 'valid_day', 'no_expiry', 'is_active'],
  })

  useEffect(() => {
    let ignore = false
    stockApi
      .getAll()
      .then((response) => {
        if (!ignore) setStocks({ status: 'ready', list: response.data ?? [] })
      })
      .catch(() => {
        if (!ignore) setStocks({ status: 'error', list: [] })
      })
    return () => {
      ignore = true
    }
  }, [])

  const markItemsChanged = () => {
    setItemsChanged(true)
    clearErrors('items') // indexes may shift, so drop every item error
  }

  const onSubmit = async (values) => {
    setSubmitFailed(false)
    const body = {
      name: values.name.trim(),
      category: values.category.trim(),
      price: String(values.price).trim(),
      max_session: Number(values.max_session),
      service_comm: String(values.service_comm || 0).trim(),
      valid_day: values.no_expiry ? null : Number(values.valid_day),
    }
    const itemsBody = values.items.map((item) => ({ stock_id: item.stock_id, use_qty: Number(item.use_qty) }))

    try {
      let saved
      if (isEdit) {
        const response = await courseApi.update(course.id, {
          ...body,
          is_active: values.is_active,
          ...(itemsChanged && { items: itemsBody }),
        })
        saved = response.data
      } else {
        // POST has no is_active (always created active), so switch it off afterwards
        const response = await courseApi.create({ ...body, items: itemsBody })
        saved = response.data
        if (!values.is_active) {
          try {
            saved = (await courseApi.update(saved.id, { is_active: false })).data
          } catch {
            toast.error('เพิ่มคอร์สแล้ว แต่ปิดขายไม่สำเร็จ', 'กดสวิตช์บนการ์ดเพื่อปิดขายอีกครั้ง')
          }
        }
      }
      toast.success('บันทึกสำเร็จ', `${isEdit ? 'แก้ไข' : 'เพิ่ม'}คอร์ส “${saved.name}” เรียบร้อยแล้ว`)
      onSaved(saved)
    } catch (error) {
      // Nothing was saved — keep the modal open with what the user typed
      mapCourseError(error).forEach(({ field, message }) => setError(field, { type: 'server', message }))
      setSubmitFailed(true)
    }
  }

  const errorCount = countErrors(errors)
  const perSession = pricePerSession(price, maxSession)
  const validDayNumber = Number(validDay)
  const expiresOn =
    noExpiry || !isWhole(validDay) || validDayNumber < 1
      ? null
      : dayjs().add(validDayNumber, 'day').locale('th').format('D MMM BBBB')

  return (
    <Modal onClose={onClose} labelledBy={titleId} fullScreenOnMobile className="sm:max-w-[1100px]">
      <form onSubmit={handleSubmit(onSubmit)} noValidate className="flex min-h-0 flex-1 flex-col">
        <header className="flex items-start justify-between gap-4 border-b border-line px-5 py-4 sm:px-6">
          <div>
            <h2 id={titleId} className="text-lg font-bold text-ink">
              {isEdit ? 'แก้ไขคอร์ส' : 'เพิ่มคอร์สใหม่'}
            </h2>
            <p className="mt-0.5 text-[13px] text-muted">
              {isEdit
                ? 'การแก้ไขมีผลกับการขายครั้งถัดไป — คอร์สที่ขายไปแล้วไม่เปลี่ยน'
                : 'กรอกข้อมูลแพ็กเกจที่ลูกค้าจะซื้อ'}
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

        <div className="min-h-0 flex-1 overflow-y-auto px-5 py-5 sm:px-6">
          {submitFailed && (
            <div
              role="alert"
              className="mb-5 flex items-start gap-3 rounded-control border border-danger/25 bg-danger-bg px-4 py-3 text-sm text-danger"
            >
              <FiAlertCircle className="mt-0.5 size-4 shrink-0" />
              <div>
                <p className="font-semibold">บันทึกไม่สำเร็จ — ยังไม่มีข้อมูลใดถูกบันทึก</p>
                <p className="mt-0.5">
                  {errorCount > 0
                    ? `แก้ไขจุดที่มีข้อความสีแดง (${errorCount} จุด) แล้วกดบันทึกอีกครั้ง`
                    : 'บันทึกไม่สำเร็จ กรุณาลองใหม่อีกครั้ง'}
                </p>
              </div>
            </div>
          )}

          <div className="grid gap-6 lg:grid-cols-2">
            <div className="space-y-5">
              <div className="grid gap-4 sm:grid-cols-[1fr_200px]">
                <Field id={`${formId}-name`} label="ชื่อคอร์ส" required error={errors.name?.message}>
                  <input
                    id={`${formId}-name`}
                    type="text"
                    placeholder="เช่น นวดน้ำมันอโรมา 10 ครั้ง"
                    aria-invalid={errors.name ? 'true' : undefined}
                    className={inputClass(errors.name)}
                    {...register('name', {
                      validate: (value) => {
                        const name = value.trim()
                        if (!name) return 'กรุณากรอกชื่อคอร์ส'
                        if (name.length > 150) return 'ชื่อคอร์สยาวได้ไม่เกิน 150 ตัวอักษร'
                        return true
                      },
                    })}
                  />
                </Field>
                <Field id={`${formId}-category`} label="หมวดหมู่" error={errors.category?.message}>
                  <input
                    id={`${formId}-category`}
                    type="text"
                    list={`${formId}-categories`}
                    placeholder="เลือกหรือพิมพ์ใหม่"
                    className={inputClass(errors.category)}
                    {...register('category', {
                      validate: (value) => value.trim().length <= 50 || 'หมวดหมู่ยาวได้ไม่เกิน 50 ตัวอักษร',
                    })}
                  />
                  <datalist id={`${formId}-categories`}>
                    {categories.map((category) => (
                      <option key={category} value={category} />
                    ))}
                  </datalist>
                </Field>
              </div>

              <div className="grid gap-4 sm:grid-cols-3">
                <Field id={`${formId}-price`} label="ราคาคอร์ส (฿)" required error={errors.price?.message}>
                  <input
                    id={`${formId}-price`}
                    type="number"
                    inputMode="decimal"
                    min={0}
                    step="0.01"
                    className={inputClass(errors.price)}
                    {...register('price', {
                      validate: (value) => {
                        if (String(value).trim() === '') return 'กรุณากรอกราคา'
                        return Number(value) >= 0 || 'ราคาต้องไม่ติดลบ'
                      },
                    })}
                  />
                </Field>
                <Field
                  id={`${formId}-sessions`}
                  label="จำนวนครั้งที่ใช้ได้"
                  required
                  hint="อย่างน้อย 1 ครั้ง"
                  error={errors.max_session?.message}
                >
                  <input
                    id={`${formId}-sessions`}
                    type="number"
                    inputMode="numeric"
                    min={1}
                    className={inputClass(errors.max_session)}
                    {...register('max_session', {
                      validate: (value) => (isWhole(value) && Number(value) >= 1) || 'ต้องเป็นจำนวนเต็มอย่างน้อย 1',
                    })}
                  />
                </Field>
                <Field
                  id={`${formId}-comm`}
                  label="ค่ามือบริการ / ครั้ง (฿)"
                  hint="จ่ายให้ผู้ให้บริการทุกครั้งที่ใช้"
                  error={errors.service_comm?.message}
                >
                  <input
                    id={`${formId}-comm`}
                    type="number"
                    inputMode="decimal"
                    min={0}
                    step="0.01"
                    className={inputClass(errors.service_comm)}
                    {...register('service_comm', {
                      validate: (value) => String(value).trim() === '' || Number(value) >= 0 || 'ค่ามือต้องไม่ติดลบ',
                    })}
                  />
                </Field>
              </div>

              <fieldset className="rounded-card border border-line p-4">
                <legend className="sr-only">อายุคอร์ส</legend>
                <label htmlFor={`${formId}-valid-day`} className="text-sm font-medium text-ink">
                  อายุคอร์ส (วัน)
                </label>
                <div className="mt-2 flex flex-wrap items-center gap-3">
                  <input
                    id={`${formId}-valid-day`}
                    type="number"
                    inputMode="numeric"
                    min={1}
                    disabled={noExpiry}
                    placeholder={noExpiry ? 'เว้นว่าง = ไม่หมดอายุ' : 'เช่น 180'}
                    aria-invalid={errors.valid_day ? 'true' : undefined}
                    className={`${inputClass(errors.valid_day)} max-w-56 flex-1 disabled:bg-surface-2`}
                    {...register('valid_day', {
                      validate: (value, values) =>
                        values.no_expiry || (isWhole(value) && Number(value) >= 1) || 'ใส่จำนวนวันอย่างน้อย 1 วัน หรือเลือกไม่หมดอายุ',
                    })}
                  />
                  <label className="flex items-center gap-2 text-sm text-ink">
                    <input
                      type="checkbox"
                      className="size-4 accent-primary"
                      {...register('no_expiry', {
                        onChange: (event) => {
                          if (event.target.checked) setValue('valid_day', '')
                          clearErrors('valid_day')
                        },
                      })}
                    />
                    ไม่หมดอายุ
                  </label>
                  <div className="ml-auto flex gap-1.5">
                    {VALID_DAY_PRESETS.map((preset) => (
                      <button
                        key={preset.days}
                        type="button"
                        onClick={() => {
                          setValue('no_expiry', false)
                          setValue('valid_day', String(preset.days))
                          clearErrors('valid_day')
                        }}
                        className="h-8 rounded-lg border border-line-strong px-2.5 text-xs text-ink transition hover:bg-surface-2"
                      >
                        {preset.label}
                      </button>
                    ))}
                  </div>
                </div>
                {errors.valid_day && <p className="mt-1.5 text-xs text-danger">{errors.valid_day.message}</p>}
                <p className="mt-3 flex items-start gap-1.5 text-xs text-muted">
                  <FiInfo className="mt-0.5 size-3.5 shrink-0" />
                  นับจากวันที่ลูกค้าซื้อ · ไม่หมดอายุ = ระบบจะไม่กำหนดวันหมดอายุให้รายการขาย
                </p>
              </fieldset>

              <div className="flex items-center justify-between gap-4 rounded-card border border-line p-4">
                <div>
                  <p className="text-sm font-medium text-ink">เปิดขายคอร์สนี้</p>
                  <p className="mt-0.5 text-xs text-muted">ปิดแล้วจะไม่แสดงตอนขาย แต่ลูกค้าที่ซื้อไว้ยังใช้ได้ตามปกติ</p>
                </div>
                <Toggle
                  checked={isActive}
                  onChange={(checked) => setValue('is_active', checked, { shouldDirty: true })}
                  label="เปิดขายคอร์สนี้"
                />
              </div>

              <dl className="grid grid-cols-3 gap-2.5">
                <Summary label="ราคาเฉลี่ยต่อครั้ง" value={perSession == null || price === '' ? '—' : formatBaht(perSession)} />
                <Summary
                  label="ค่ามือรวมทั้งคอร์ส"
                  value={formatBaht(Math.round(toNumber(serviceComm) * toNumber(maxSession)))}
                />
                <Summary label="ถ้าขายวันนี้ หมดอายุ" value={expiresOn ?? 'ไม่หมดอายุ'} />
              </dl>
            </div>

            <CourseItemsField
              rows={fields}
              values={items}
              stocks={stocks}
              errors={errors.items}
              changed={itemsChanged}
              onAdd={(stock) => {
                append({ stock_id: stock.id, stock_name: stock.name, use_qty: 1 })
                markItemsChanged()
              }}
              onRemove={(index) => {
                remove(index)
                markItemsChanged()
              }}
              onQtyChange={(index, qty) => {
                setValue(`items.${index}.use_qty`, qty)
                setItemsChanged(true)
                clearErrors(`items.${index}`)
              }}
            />
          </div>
        </div>

        <footer className="flex justify-end gap-3 border-t border-line px-5 py-4 sm:px-6">
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
            {isSubmitting ? 'กำลังบันทึก…' : isEdit ? 'บันทึกการแก้ไข' : 'เพิ่มคอร์ส'}
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

function Field({ id, label, required, hint, error, children }) {
  return (
    <div>
      <label htmlFor={id} className="mb-1.5 block text-sm font-medium text-ink">
        {label}
        {required && <span className="ml-0.5 text-danger">*</span>}
      </label>
      {children}
      {error ? (
        <p className="mt-1.5 text-xs text-danger">{error}</p>
      ) : (
        hint && <p className="mt-1.5 text-xs text-muted">{hint}</p>
      )}
    </div>
  )
}

function Summary({ label, value }) {
  return (
    <div className="min-w-0 rounded-control bg-gold-bg px-3 py-2.5">
      <dt className="text-[11px] text-gold-ink">{label}</dt>
      <dd className="mt-1 text-[15px] font-bold text-ink [overflow-wrap:anywhere]">{value}</dd>
    </div>
  )
}

export default CourseFormModal
