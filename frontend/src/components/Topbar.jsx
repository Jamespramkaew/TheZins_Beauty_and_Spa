import { useEffect, useRef } from 'react'
import { Link, useMatches } from 'react-router-dom'
import dayjs from 'dayjs'
import buddhistEra from 'dayjs/plugin/buddhistEra'
import 'dayjs/locale/th'
import { LuBell, LuCalendar, LuChevronRight, LuFlower2, LuMenu, LuPlus, LuSearch } from 'react-icons/lu'

dayjs.extend(buddhistEra)

function Topbar({ onOpenMobileMenu }) {
  const searchRef = useRef(null)
  const handle = useMatches().findLast((match) => match.handle)?.handle ?? {}
  const today = dayjs().locale('th').format('[วัน]dddd[ที่] D MMM BBBB')

  useEffect(() => {
    const onKeyDown = (event) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault()
        searchRef.current?.focus()
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [])

  return (
    <header className="sticky top-0 z-20 flex h-18 items-center short:h-16 gap-3 border-b border-[#EEEDF0] bg-white/90 px-4 backdrop-blur sm:px-6 lg:px-8">
      <button
        type="button"
        onClick={onOpenMobileMenu}
        aria-label="เปิดเมนู"
        className="flex size-10 items-center justify-center rounded-control border border-line-strong text-muted lg:hidden"
      >
        <LuMenu className="size-5" />
      </button>

      <nav aria-label="Breadcrumb" className="flex min-w-0 items-center gap-2 text-sm">
        <span className="hidden items-center gap-1.5 rounded-lg bg-gold-bg px-2.5 py-1 font-medium text-gold-ink sm:flex">
          <LuFlower2 className="size-3.5 text-gold" />
          The Zins
        </span>
        <LuChevronRight className="hidden size-4 text-subtle sm:block" />
        <span className="truncate font-semibold text-ink">{handle.title}</span>
      </nav>

      <label className="relative ml-2 hidden max-w-md flex-1 md:block">
        <LuSearch className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-subtle" />
        <input
          ref={searchRef}
          type="search"
          placeholder="ค้นหาลูกค้า, ออเดอร์, คอร์ส..."
          className="h-10 w-full rounded-control border border-line-strong bg-surface pr-12 pl-10 text-sm placeholder:text-subtle focus:border-chart/40 focus:bg-white focus:ring-3 focus:ring-chart/10 focus:outline-none"
        />
        <kbd className="pointer-events-none absolute top-1/2 right-3 -translate-y-1/2 rounded border border-line-strong bg-white px-1.5 font-sans text-[11px] text-subtle">
          ⌘K
        </kbd>
      </label>

      <div className="ml-auto flex items-center gap-3">
        <span className="hidden items-center gap-2 text-sm text-muted xl:flex">
          <LuCalendar className="size-4" />
          {today}
        </span>
        <button
          type="button"
          aria-label="การแจ้งเตือน"
          className="relative flex size-10 items-center justify-center rounded-control border border-line-strong text-muted transition hover:bg-surface"
        >
          <LuBell className="size-4.5" />
          <span className="absolute top-2 right-2.5 size-2 rounded-full bg-[#D92D20] ring-2 ring-white" />
        </button>
        {handle.showCreateOrder && (
          <Link
            to="/sell"
            className="flex h-10 items-center gap-2 rounded-control bg-primary px-4 text-sm font-semibold text-white shadow-sm transition hover:bg-[#4d1621]"
          >
            <LuPlus className="size-4" />
            <span className="hidden sm:inline">สร้างออเดอร์</span>
          </Link>
        )}
      </div>
    </header>
  )
}

export default Topbar
