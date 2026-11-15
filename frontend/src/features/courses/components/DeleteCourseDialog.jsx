import { useId, useState } from 'react'
import { FiInfo, FiTrash2 } from 'react-icons/fi'
import Modal from '../../../components/Modal'
import { courseApi } from '../../../api/courseApi'
import { useToast } from '../../../hooks/useToast'

// Backend decides hard vs soft delete (sold courses keep their history); both look the same here.
function DeleteCourseDialog({ course, onClose, onDeleted, onDeactivated }) {
  const toast = useToast()
  const titleId = useId()
  const descriptionId = useId()
  const [pending, setPending] = useState(null) // 'delete' | 'deactivate'

  const deleteCourse = async () => {
    setPending('delete')
    try {
      await courseApi.remove(course.id)
      toast.success('ลบคอร์สแล้ว', `ลบ “${course.name}” ออกจากรายการเรียบร้อยแล้ว`)
      onDeleted(course)
    } catch {
      toast.error('ลบคอร์สไม่สำเร็จ', 'กรุณาลองใหม่อีกครั้ง')
      setPending(null)
    }
  }

  const deactivate = async () => {
    setPending('deactivate')
    try {
      const response = await courseApi.update(course.id, { is_active: false })
      toast.success('ปิดขายแล้ว', `“${course.name}” ยังอยู่ในรายการ และเปิดขายกลับได้ทุกเมื่อ`)
      onDeactivated(response.data)
    } catch {
      toast.error('ปิดขายไม่สำเร็จ', 'กรุณาลองใหม่อีกครั้ง')
      setPending(null)
    }
  }

  return (
    <Modal role="alertdialog" onClose={onClose} labelledBy={titleId} describedBy={descriptionId} className="max-w-[480px] p-6">
      <span className="flex size-11 items-center justify-center rounded-control bg-danger-bg text-danger">
        <FiTrash2 className="size-5" />
      </span>
      <h2 id={titleId} className="mt-4 text-lg font-bold text-ink [overflow-wrap:anywhere]">
        ลบ “{course.name}” ?
      </h2>
      <p id={descriptionId} className="mt-1.5 text-sm text-muted">
        คอร์สจะหายจากรายการ <span className="font-semibold text-ink">ประวัติการขายเดิมยังอยู่ครบ</span>
      </p>

      {course.is_active && (
        <p className="mt-4 flex items-start gap-2.5 rounded-control bg-gold-bg px-4 py-3 text-[13px] text-gold-ink">
          <FiInfo className="mt-0.5 size-4 shrink-0" />
          <span>
            ถ้าแค่อยากหยุดขายชั่วคราว ใช้ <span className="font-semibold">“ปิดขาย”</span> แทน —
            คอร์สยังอยู่ในรายการและเปิดขายกลับได้ทุกเมื่อ
          </span>
        </p>
      )}

      <div className="mt-6 flex flex-wrap justify-end gap-2.5">
        <button
          type="button"
          onClick={onClose}
          className="h-10 rounded-control border border-line-strong px-4 text-sm font-medium text-ink transition hover:bg-surface-2"
        >
          ยกเลิก
        </button>
        {course.is_active && (
          <button
            type="button"
            onClick={deactivate}
            disabled={pending !== null}
            className="h-10 rounded-control border border-line-strong px-4 text-sm font-semibold text-ink transition hover:bg-surface-2 disabled:opacity-60"
          >
            {pending === 'deactivate' ? 'กำลังปิดขาย…' : 'ปิดขายแทน'}
          </button>
        )}
        <button
          type="button"
          onClick={deleteCourse}
          disabled={pending !== null}
          className="h-10 rounded-control bg-[#9B1C1C] px-4 text-sm font-semibold text-white transition hover:bg-[#831717] disabled:opacity-60"
        >
          {pending === 'delete' ? 'กำลังลบ…' : 'ลบคอร์ส'}
        </button>
      </div>
    </Modal>
  )
}

export default DeleteCourseDialog
