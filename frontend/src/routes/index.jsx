import { createBrowserRouter } from 'react-router-dom'
import MainLayout from '../layouts/MainLayout'
import HomePage from '../pages/HomePage'
import ServicesPage from '../pages/ServicesPage'
import BookingPage from '../pages/BookingPage'
import NotFoundPage from '../pages/NotFoundPage'

export const router = createBrowserRouter([
  {
    path: '/',
    element: <MainLayout />,
    children: [
      { index: true, element: <HomePage /> },
      { path: 'services', element: <ServicesPage /> },
      { path: 'booking', element: <BookingPage /> },
      { path: '*', element: <NotFoundPage /> },
    ],
  },
])
