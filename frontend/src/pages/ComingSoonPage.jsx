import { useMatches } from 'react-router-dom'
import { LuHammer } from 'react-icons/lu'

function ComingSoonPage() {
  const title = useMatches().findLast((match) => match.handle)?.handle.title

  return (
    <div className="mx-auto max-w-[1400px]">
      <h1 className="text-[28px] font-bold tracking-tight text-ink">{title}</h1>
      <TempToastDemo />
      <div className="mt-6 flex flex-col items-center justify-center rounded-card border border-dashed border-line-strong bg-surface px-6 py-20 text-center">
        <span className="flex size-12 items-center justify-center rounded-full bg-gold-bg text-gold-ink">
          <LuHammer className="size-5" />
        </span>
        <p className="mt-4 font-semibold text-ink">หน้านี้กำลังพัฒนา</p>
        <p className="mt-1 text-sm text-muted">เร็ว ๆ นี้</p>
      </div>
    </div>
  )
}

export default ComingSoonPage

// TEMP-DEMO-START — toast preview only, remove before commit
import { useToast } from '../hooks/useToast'
function TempToastDemo() {
  const toast = useToast()
  return (
    <div className="mt-4 flex gap-3">
      <button type="button" onClick={() => toast.success('บันทึกสำเร็จ', 'เพิ่มคอร์ส “หน้าใส Aura” เรียบร้อยแล้ว')} className="h-10 rounded-control bg-primary px-4 text-sm font-semibold text-white">
        ทดสอบ Success
      </button>
      <button type="button" onClick={() => toast.error('บันทึกไม่สำเร็จ', 'จำนวนที่โอนเกินยอดคงเหลือของสาขาต้นทาง (คงเหลือ 3 ชิ้น) กรุณาตรวจสอบจำนวนสินค้าและสาขาปลายทางอีกครั้งก่อนบันทึกรายการ')} className="h-10 rounded-control border border-line-strong px-4 text-sm font-semibold text-ink">
        ทดสอบ Fail
      </button>
    </div>
  )
}
// TEMP-DEMO-END
