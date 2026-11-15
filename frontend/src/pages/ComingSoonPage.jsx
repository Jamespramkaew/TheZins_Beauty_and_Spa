import { useMatches } from 'react-router-dom'
import { LuHammer } from 'react-icons/lu'

function ComingSoonPage() {
  const title = useMatches().findLast((match) => match.handle)?.handle.title

  return (
    <div className="mx-auto max-w-[1400px]">
      <h1 className="text-[28px] font-bold tracking-tight text-ink">{title}</h1>
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
