import { NavLink, Outlet } from 'react-router-dom'
import { env } from '../config/env'

const menus = [
  { to: '/', label: 'หน้าแรก' },
  { to: '/services', label: 'บริการ' },
  { to: '/booking', label: 'จองคิว' },
]

function MainLayout() {
  return (
    <div className="flex min-h-screen flex-col">
      <header className="border-b px-6 py-4">
        <nav className="flex items-center gap-6">
          <span className="font-bold text-pink-500">{env.appName}</span>
          {menus.map((menu) => (
            <NavLink
              key={menu.to}
              to={menu.to}
              end
              className={({ isActive }) =>
                isActive ? 'font-semibold text-pink-500' : 'text-gray-600'
              }
            >
              {menu.label}
            </NavLink>
          ))}
        </nav>
      </header>

      <main className="flex-1 p-6">
        <Outlet />
      </main>

      <footer className="border-t px-6 py-4 text-center text-sm text-gray-500">
        © {new Date().getFullYear()} {env.appName}
      </footer>
    </div>
  )
}

export default MainLayout
