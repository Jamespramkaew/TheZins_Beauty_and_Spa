import { createBrowserRouter } from 'react-router-dom'
import DashboardLayout from '../layouts/DashboardLayout'
import ComingSoonPage from '../pages/ComingSoonPage'
import NotFoundPage from '../pages/NotFoundPage'
import OverviewPage from '../features/overview/OverviewPage'
import CoursesPage from '../features/courses/CoursesPage'
import StocksPage from '../features/stocks/StocksPage'
import LoginPage from '../features/auth/LoginPage'
import RequireAuth from '../features/auth/RequireAuth'

export const router = createBrowserRouter([
  // Login is the entry page of the system
  { path: '/', element: <LoginPage /> },
  {
    // Back-office routes mirror the Sidebar menu and require login. Replace ComingSoonPage
    // with the module's page once it's built. `handle.title` feeds the Topbar breadcrumb.
    element: (
      <RequireAuth>
        <DashboardLayout />
      </RequireAuth>
    ),
    children: [
      {
        path: 'overview',
        element: <OverviewPage />,
        handle: { title: 'Overview', showCreateOrder: true },
      },
      { path: 'orders', element: <ComingSoonPage />, handle: { title: 'Orders' } },
      { path: 'customers', element: <ComingSoonPage />, handle: { title: 'Customers' } },
      { path: 'stocks', element: <StocksPage />, handle: { title: 'Stocks & Stock-Logs' } },
      { path: 'courses', element: <CoursesPage />, handle: { title: 'Courses' } },
      { path: 'employees', element: <ComingSoonPage />, handle: { title: 'Employee' } },
      { path: 'branch-accounts', element: <ComingSoonPage />, handle: { title: 'Branch Account' } },
      { path: '*', element: <NotFoundPage />, handle: { title: 'ไม่พบหน้านี้' } },
    ],
  },
])
