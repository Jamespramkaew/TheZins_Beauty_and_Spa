import { createBrowserRouter, Navigate } from 'react-router-dom'
import DashboardLayout from '../layouts/DashboardLayout'
import ComingSoonPage from '../pages/ComingSoonPage'
import NotFoundPage from '../pages/NotFoundPage'
import OverviewPage from '../features/overview/OverviewPage'

// Routes mirror the Sidebar menu. Replace ComingSoonPage with the module's page once it's built.
// `handle.title` feeds the Topbar breadcrumb.
export const router = createBrowserRouter([
  {
    path: '/',
    element: <DashboardLayout />,
    children: [
      { index: true, element: <Navigate to="/overview" replace /> },
      {
        path: 'overview',
        element: <OverviewPage />,
        handle: { title: 'Overview', showCreateOrder: true },
      },
      { path: 'orders', element: <ComingSoonPage />, handle: { title: 'Orders' } },
      { path: 'customers', element: <ComingSoonPage />, handle: { title: 'Customers' } },
      { path: 'stocks', element: <ComingSoonPage />, handle: { title: 'Stocks & Stock-Logs' } },
      { path: 'courses', element: <ComingSoonPage />, handle: { title: 'Courses' } },
      { path: 'employees', element: <ComingSoonPage />, handle: { title: 'Employee' } },
      { path: 'branch-accounts', element: <ComingSoonPage />, handle: { title: 'Branch Account' } },
      { path: '*', element: <NotFoundPage />, handle: { title: 'ไม่พบหน้านี้' } },
    ],
  },
])
