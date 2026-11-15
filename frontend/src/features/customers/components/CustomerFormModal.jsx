import { useId, useState } from 'react'
import { useForm } from 'react-hook-form'
import dayjs from 'dayjs'
import { FiAlertCircle, FiX } from 'react-icons/fi'
import Modal from '../../../components/Modal'
import { customerApi } from '../../../api/customerApi'
import { useToast } from '../../../hooks/useToast'
import { mapCustomerError } from '../customerErrors'
import { fullName } from '../customerUtils'

// b_day comes back as "2000-01-31T00:00:00Z"; <input type="date"> wants "2000-01-31"
const toFormValues = (customer) => ({
  hn_no: customer?.hn_no ?? '',
  fname: customer?.fname ?? '',
  lname: customer?.lname ?? '',
  tel: customer?.tel ?? '',
  citizen_id: customer?.citizen_id ?? '',
  b_day: customer?.b_day ? customer.b_day.slice(0, 10) : '',
  address: customer?.address ?? '',
})

// Same limits as the backend (internal/customer/service.go)
const maxLength = (label, max) => (value) => value.trim().length <= max || `${label}ยาวได้ไม่เกิน ${max} ตัวอักษร`

// customer = null → create
function CustomerFormModal({ customer, onClose, onSaved }) {
  const isEdit = Boolean(customer)
  const toast = useToast()
  const formId = useId()
  const titleId = `${formId}-title`
  const [submitFailed, setSubmitFailed] = useState(false)
  const today = dayjs().format('YYYY-MM-DD')

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm({ defaultValues: toFormValues(customer) })

  const onSubmit = async (values) => {
    setSubmitFailed(false)
    // Empty citizen_id / b_day = no value (NULL); on edit that also clears what was there
    const body = {
      hn_no: values.hn_no.trim(),
      fname: values.fname.trim(),
      lname: values.lname.trim(),
      tel: values.tel.trim(),
      citizen_id: values.citizen_id.trim(),
      b_day: values.b_day,
      address: values.address.trim(),
    }

    try {
      const response = isEdit ? await customerApi.update(customer.id, body) : await customerApi.create(body)
      toast.success('บันทึกสำเร็จ', `${isEdit ? 'แก้ไข' : 'เพิ่ม'}ลูกค้า “${fullName(response.data)}” เรียบร้อยแล้ว`)
      onSaved(response.data)
    } catch (error) {
      // Nothing was saved — keep the modal open with what the user typed
      mapCustomerError(error).forEach(({ field, message }) => setError(field, { type: 'server', message }))
      setSubmitFailed(true)
    }
  }

  const errorCount = Object.keys(errors).length
  const fieldId = (name) => `${formId}-${name}`

  return (
    <Modal onClose={onClose} labelledBy={titleId} fullScreenOnMobile className="sm:max-w-[720px]">
      <form onSubmit={handleSubmit(onSubmit)} noValidate className="flex min-h-0 flex-1 flex-col">
        <header className="flex items-start justify-between gap-4 border-b border-line px-5 py-4 sm:px-6">
          <div>
            <h2 id={titleId} className="text-lg font-bold text-ink">
              {isEdit ? 'แก้ไขข้อมูลลูกค้า' : 'เพิ่มลูกค้าใหม่'}
            </h2>
            <p className="mt-0.5 text-[13px] text-muted">
              {isEdit ? 'ระดับลูกค้า (VIP / ทั่วไป) เปลี่ยนได้จากรายละเอียดลูกค้า' : 'ลูกค้าใหม่เริ่มที่ระดับทั่วไป'}
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

          <div className="grid gap-4 sm:grid-cols-2">
            <Field id={fieldId('hn_no')} label="เลข HN" required hint="เลขประจำตัวลูกค้าที่ร้านกำหนด ห้ามซ้ำ" error={errors.hn_no?.message}>
              <input
                id={fieldId('hn_no')}
                type="text"
                placeholder="เช่น HN000128"
                aria-invalid={errors.hn_no ? 'true' : undefined}
                className={inputClass(errors.hn_no)}
                {...register('hn_no', {
                  validate: (value) => (value.trim() ? maxLength('เลข HN', 20)(value) : 'กรุณากรอกเลข HN'),
                })}
              />
            </Field>
            <Field id={fieldId('tel')} label="เบอร์โทร" error={errors.tel?.message}>
              <input
                id={fieldId('tel')}
                type="tel"
                inputMode="tel"
                placeholder="เช่น 0891234567"
                aria-invalid={errors.tel ? 'true' : undefined}
                className={inputClass(errors.tel)}
                {...register('tel', { validate: maxLength('เบอร์โทร', 20) })}
              />
            </Field>
            <Field id={fieldId('fname')} label="ชื่อ" required error={errors.fname?.message}>
              <input
                id={fieldId('fname')}
                type="text"
                aria-invalid={errors.fname ? 'true' : undefined}
                className={inputClass(errors.fname)}
                {...register('fname', {
                  validate: (value) => (value.trim() ? maxLength('ชื่อ', 100)(value) : 'กรุณากรอกชื่อ'),
                })}
              />
            </Field>
            <Field id={fieldId('lname')} label="นามสกุล" error={errors.lname?.message}>
              <input
                id={fieldId('lname')}
                type="text"
                aria-invalid={errors.lname ? 'true' : undefined}
                className={inputClass(errors.lname)}
                {...register('lname', { validate: maxLength('นามสกุล', 100) })}
              />
            </Field>
            <Field id={fieldId('citizen_id')} label="เลขบัตรประชาชน" error={errors.citizen_id?.message}>
              <input
                id={fieldId('citizen_id')}
                type="text"
                inputMode="numeric"
                maxLength={13}
                placeholder="ตัวเลข 13 หลัก"
                aria-invalid={errors.citizen_id ? 'true' : undefined}
                className={inputClass(errors.citizen_id)}
                {...register('citizen_id', {
                  validate: (value) =>
                    !value.trim() || /^[0-9]{13}$/.test(value.trim()) || 'เลขบัตรประชาชนต้องเป็นตัวเลข 13 หลัก',
                })}
              />
            </Field>
            <Field id={fieldId('b_day')} label="วันเกิด" error={errors.b_day?.message}>
              <input
                id={fieldId('b_day')}
                type="date"
                max={today}
                aria-invalid={errors.b_day ? 'true' : undefined}
                className={inputClass(errors.b_day)}
                {...register('b_day', {
                  validate: (value) => !value || value <= today || 'วันเกิดต้องไม่เป็นวันในอนาคต',
                })}
              />
            </Field>
            <div className="sm:col-span-2">
              <Field id={fieldId('address')} label="ที่อยู่" error={errors.address?.message}>
                <textarea
                  id={fieldId('address')}
                  rows={3}
                  className={`${inputClass(errors.address)} h-auto py-2.5`}
                  {...register('address')}
                />
              </Field>
            </div>
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
            {isSubmitting ? 'กำลังบันทึก…' : isEdit ? 'บันทึกการแก้ไข' : 'เพิ่มลูกค้า'}
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

export default CustomerFormModal
