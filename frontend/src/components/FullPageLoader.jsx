import { LuLoaderCircle } from 'react-icons/lu'

function FullPageLoader() {
  return (
    <div role="status" className="flex min-h-screen flex-col items-center justify-center gap-4 bg-white">
      <img src="/thezin-logo.png" alt="" className="h-14 w-auto" />
      <LuLoaderCircle className="size-6 animate-spin text-gold" aria-hidden="true" />
      <span className="sr-only">กำลังโหลด…</span>
    </div>
  )
}

export default FullPageLoader
