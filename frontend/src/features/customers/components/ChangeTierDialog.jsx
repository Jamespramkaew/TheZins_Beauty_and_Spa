import { useId, useState } from 'react'
import { FiStar } from 'react-icons/fi'
import Modal from '../../../components/Modal'
import { customerApi } from '../../../api/customerApi'
import { useToast } from '../../../hooks/useToast'
import { TIER_VALUES, fullName, tierOf } from '../customerUtils'
import TierBadge from './TierBadge'

// Pick any tier other than the current one. The backend records who changed it and when.
function ChangeTierDialog({ customer, onClose, onChanged }) {
  const toast = useToast()
  const titleId = useId()
  const descriptionId = useId()
  const options = TIER_VALUES.filter((tier) => tier !== customer.tier)
  const [nextTier, setNextTier] = useState(options[0])
  const [pending, setPending] = useState(false)
  const name = fullName(customer)

  const changeTier = async () => {
    setPending(true)
    try {
      const response = await customerApi.changeTier(customer.id, nextTier)
      toast.success('เปลี่ยนระดับแล้ว', `“${name}” เป็น ${tierOf(nextTier).label} แล้ว`)
      onChanged(response.data)
    } catch {
      toast.error('เปลี่ยนระดับไม่สำเร็จ', 'กรุณาลองใหม่อีกครั้ง')
      setPending(false)
    }
  }

  return (
    <Modal role="alertdialog" onClose={onClose} labelledBy={titleId} describedBy={descriptionId} className="max-w-[480px] p-6">
      <span className="flex size-11 items-center justify-center rounded-control bg-gold-bg text-gold-ink">
        <FiStar className="size-5" />
      </span>
      <h2 id={titleId} className="mt-4 text-lg font-bold text-ink [overflow-wrap:anywhere]">
        เปลี่ยนระดับของ “{name}”
      </h2>
      <p id={descriptionId} className="mt-1.5 text-sm text-muted">
        ตอนนี้เป็น <span className="font-semibold text-ink">{tierOf(customer.tier).label}</span> — เลือกระดับใหม่
        ระบบเก็บประวัติว่าใครเปลี่ยนและเมื่อไหร่
      </p>

      <fieldset className="mt-4 space-y-2">
        <legend className="sr-only">ระดับใหม่</legend>
        {options.map((tier) => (
          <label
            key={tier}
            className={`flex cursor-pointer items-center gap-3 rounded-control border px-4 py-3 transition ${
              nextTier === tier ? 'border-primary bg-surface-2' : 'border-line-strong hover:bg-surface-2'
            }`}
          >
            <input
              type="radio"
              name={`${titleId}-tier`}
              value={tier}
              checked={nextTier === tier}
              onChange={() => setNextTier(tier)}
              className="size-4 accent-primary"
            />
            <TierBadge tier={tier} />
          </label>
        ))}
      </fieldset>

      <div className="mt-6 flex flex-wrap justify-end gap-2.5">
        <button
          type="button"
          onClick={onClose}
          className="h-10 rounded-control border border-line-strong px-4 text-sm font-medium text-ink transition hover:bg-surface-2"
        >
          ยกเลิก
        </button>
        <button
          type="button"
          onClick={changeTier}
          disabled={pending}
          className="h-10 rounded-control bg-primary px-4 text-sm font-semibold text-white transition hover:bg-[#4d1621] disabled:opacity-60"
        >
          {pending ? 'กำลังเปลี่ยน…' : `เปลี่ยนเป็น ${tierOf(nextTier).label}`}
        </button>
      </div>
    </Modal>
  )
}

export default ChangeTierDialog
