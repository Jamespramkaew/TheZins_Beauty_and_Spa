import { NavLink } from 'react-router-dom'
import {
  LuChevronDown,
  LuClipboardList,
  LuIdCard,
  LuLayoutGrid,
  LuLogOut,
  LuPackage,
  LuPanelLeft,
  LuSparkles,
  LuStore,
  LuUsers,
  LuX,
} from 'react-icons/lu'
import { currentUser } from '../mocks/session'

// TODO: stocks badge should be the real count of low/out-of-stock items
const menuGroups = [
  {
    label: 'GENERAL',
    items: [
      { to: '/overview', label: 'Overview', icon: LuLayoutGrid },
      { to: '/orders', label: 'Orders', icon: LuClipboardList },
      { to: '/customers', label: 'Customers', icon: LuUsers },
      { to: '/stocks', label: 'Stocks & Stock-Logs', icon: LuPackage, badge: 2 },
      { to: '/courses', label: 'Courses', icon: LuSparkles },
    ],
  },
  {
    label: 'ADVANCE',
    items: [
      { to: '/employees', label: 'Employee', icon: LuIdCard },
      { to: '/branch-accounts', label: 'Branch Account', icon: LuStore, adminOnly: true },
    ],
  },
]

function Sidebar({ collapsed, mobileOpen, onToggleCollapse, onCloseMobile }) {
  const isAdmin = currentUser.role === 'admin'

  return (
    <>
      {mobileOpen && (
        <div
          className="fixed inset-0 z-30 bg-black/40 lg:hidden"
          onClick={onCloseMobile}
          aria-hidden="true"
        />
      )}

      <aside
        className={`fixed inset-y-0 left-0 z-40 flex w-62 flex-col bg-sidebar text-sidebar-text transition-[width,translate] duration-200 lg:sticky lg:top-0 lg:h-screen lg:translate-x-0 ${
          mobileOpen ? 'translate-x-0' : '-translate-x-full'
        } ${collapsed ? 'lg:w-19' : 'lg:w-62'}`}
      >
        <div
          className={`flex h-18 shrink-0 items-center border-b border-white/10 px-4 short:h-16 ${
            collapsed ? 'lg:justify-center' : 'justify-between'
          }`}
        >
          <img
            src="/thezin-logo.png"
            alt="The Zin Beauty & Spa"
            className={`h-11 w-auto ${collapsed ? 'lg:hidden' : ''}`}
          />
          <button
            type="button"
            onClick={onToggleCollapse}
            aria-label={collapsed ? 'ขยายเมนู' : 'ย่อเมนู'}
            className="hidden size-9 items-center justify-center rounded-control text-sidebar-text/70 transition hover:bg-white/10 hover:text-white lg:flex"
          >
            <LuPanelLeft className="size-4.5" />
          </button>
          <button
            type="button"
            onClick={onCloseMobile}
            aria-label="ปิดเมนู"
            className="flex size-9 items-center justify-center rounded-control text-sidebar-text/70 hover:bg-white/10 lg:hidden"
          >
            <LuX className="size-5" />
          </button>
        </div>

        <nav className="scrollbar-thin-dark flex-1 space-y-6 overflow-y-auto px-3 py-6 short:space-y-4 short:py-4">
          {menuGroups.map((group) => (
            <div key={group.label}>
              <p
                className={`mb-2 flex items-center gap-1 px-3 text-[11px] short:mb-1 font-semibold tracking-[0.14em] text-gold ${
                  collapsed ? 'lg:justify-center lg:px-0' : ''
                }`}
              >
                <span className={collapsed ? 'lg:hidden' : ''}>{group.label}</span>
                <LuChevronDown className={`size-3 ${collapsed ? 'lg:hidden' : ''}`} />
                {collapsed && <span className="hidden h-px w-6 bg-gold/40 lg:block" />}
              </p>

              <ul className="space-y-1">
                {group.items
                  .filter((item) => !item.adminOnly || isAdmin)
                  .map((item) => (
                    <li key={item.to}>
                      <SidebarLink item={item} collapsed={collapsed} onNavigate={onCloseMobile} />
                    </li>
                  ))}
              </ul>
            </div>
          ))}
        </nav>

        <div className="shrink-0 p-3 short:pt-2">
          <div
            className={`flex items-center gap-3 rounded-card border border-white/10 bg-white/5 p-3 short:p-2.5 ${
              collapsed ? 'lg:flex-col lg:p-2' : ''
            }`}
          >
            <span className="flex size-9 shrink-0 items-center justify-center rounded-full bg-gold text-sm font-bold text-sidebar">
              {currentUser.initial}
            </span>
            <div className={`min-w-0 flex-1 ${collapsed ? 'lg:hidden' : ''}`}>
              <p className="truncate text-sm font-semibold text-white">{currentUser.name}</p>
              <p className="truncate text-xs text-sidebar-text/70">
                {currentUser.roleLabel} · {currentUser.branchLabel}
              </p>
            </div>
            <button
              type="button"
              aria-label="ออกจากระบบ"
              title="ออกจากระบบ"
              className="flex size-8 shrink-0 items-center justify-center rounded-lg text-sidebar-text/70 transition hover:bg-white/10 hover:text-white"
            >
              <LuLogOut className="size-4" />
            </button>
          </div>
        </div>
      </aside>
    </>
  )
}

function SidebarLink({ item, collapsed, onNavigate }) {
  const Icon = item.icon

  return (
    <NavLink
      to={item.to}
      onClick={onNavigate}
      title={collapsed ? item.label : undefined}
      className={({ isActive }) =>
        `group relative flex min-h-10 items-center gap-3 rounded-control px-3 py-2 text-sm short:min-h-9 short:py-1.5 font-medium transition ${
          isActive ? 'bg-white/9 text-white' : 'text-sidebar-text hover:bg-white/5 hover:text-white'
        } ${collapsed ? 'lg:justify-center lg:px-0' : ''}`
      }
    >
      {({ isActive }) => (
        <>
          <span
            className={`flex size-7 shrink-0 items-center justify-center rounded-md transition ${
              isActive ? 'bg-gold text-sidebar' : 'text-sidebar-text/80 group-hover:text-white'
            }`}
          >
            <Icon className="size-4" />
          </span>
          <span className={`flex-1 truncate ${collapsed ? 'lg:hidden' : ''}`}>{item.label}</span>
          {item.badge > 0 && (
            <span
              className={`flex size-5 items-center justify-center rounded-full bg-white text-[11px] font-bold text-sidebar ${
                collapsed ? 'lg:absolute lg:top-0.5 lg:right-1.5 lg:size-4 lg:text-[10px]' : ''
              }`}
            >
              {item.badge}
            </span>
          )}
        </>
      )}
    </NavLink>
  )
}

export default Sidebar
