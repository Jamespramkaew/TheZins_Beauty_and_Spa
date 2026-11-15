import { useId, useState } from 'react'
import { FiInfo, FiTrash2 } from 'react-icons/fi'
import Modal from '../../../components/Modal'
import { stockApi } from '../../../api/stockApi'
import { useToast } from '../../../hooks/useToast'
import { outlineButtonClass } from './styles'

// Backend decides: products with history are hidden (soft delete), products never used are removed for good.
function DeleteStockDialog({ stock, onClose, onDeleted, onDeactivated }) {
  const toast = useToast()
  const titleId = useId()
  const descriptionId = useId()
  const [pending, setPending] = useState(null) // 'delete' | 'deactivate'

  const deleteStock = async () => {
    setPending('delete')
    try {
      await stockApi.remove(stock.id)
      toast.success('ลบสินค้าแล้ว', `ลบ “${stock.name}” ออกจากรายการเรียบร้อยแล้ว`)
      onDeleted(stock)
    } catch {
      toast.error('ลบสินค้าไม่สำเร็จ', 'กรุณาลองใหม่อีกครั้ง')
      setPending(null)
    }
  }

  const deactivate = async () => {
    setPending('deactivate')
    try {
      const response = await stockApi.update(stock.id, { is_active: false })
      toast.success('ปิดใช้งานแล้ว', `“${stock.name}” ยังอยู่ในรายการ และเปิดใช้งานกลับได้ทุกเมื่อ`)
      onDeactivated(response.data)
    } catch {
      toast.error('ปิดใช้งานไม่สำเร็จ', 'กรุณาลองใหม่อีกครั้ง')
      setPending(null)
    }
  }

  return (
    <Modal role="alertdialog" onClose={onClose} labelledBy={titleId} describedBy={descriptionId} className="max-w-[480px] p-6">
      <span className="flex size-11 items-center justify-center rounded-control bg-danger-bg text-danger">
        <FiTrash2 className="size-5" />
      </span>
      <h2 id={titleId} className="mt-4 text-lg font-bold text-ink [overflow-wrap:anywhere]">
        ลบ “{stock.name}” ?
      </h2>
      <p id={descriptionId} className="mt-1.5 text-sm text-muted">
        ถ้าสินค้านี้เคยมีความเคลื่อนไหว ระบบจะซ่อนไว้และ <span className="font-semibold text-ink">ประวัติเดิมยังอยู่ครบ</span>{' '}
        ถ้ายังไม่เคยใช้เลย จะถูกลบถาวร
      </p>

      {stock.is_active && (
        <p className="mt-4 flex items-start gap-2.5 rounded-control bg-gold-bg px-4 py-3 text-[13px] text-gold-ink">
          <FiInfo className="mt-0.5 size-4 shrink-0" />
          <span>
            ถ้าแค่อยากหยุดใช้ชั่วคราว ใช้ <span className="font-semibold">“ปิดใช้งาน”</span> แทน — สินค้ายังอยู่ในรายการและเปิดกลับได้ทุกเมื่อ
          </span>
        </p>
      )}

      <div className="mt-6 flex flex-wrap justify-end gap-2.5">
        <button type="button" onClick={onClose} className={`${outlineButtonClass} h-10`}>
          ยกเลิก
        </button>
        {stock.is_active && (
          <button
            type="button"
            onClick={deactivate}
            disabled={pending !== null}
            className={`${outlineButtonClass} h-10 font-semibold`}
          >
            {pending === 'deactivate' ? 'กำลังปิดใช้งาน…' : 'ปิดใช้งานแทน'}
          </button>
        )}
        <button
          type="button"
          onClick={deleteStock}
          disabled={pending !== null}
          className="h-10 rounded-control bg-[#9B1C1C] px-4 text-sm font-semibold text-white transition hover:bg-[#831717] focus-visible:ring-3 focus-visible:ring-danger/30 focus-visible:outline-none disabled:opacity-60"
        >
          {pending === 'delete' ? 'กำลังลบ…' : 'ลบสินค้า'}
        </button>
      </div>
    </Modal>
  )
}

export default DeleteStockDialog
