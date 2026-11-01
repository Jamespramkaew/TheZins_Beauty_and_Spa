import { Link } from 'react-router-dom'

function NotFoundPage() {
  return (
    <div className="flex flex-col items-center py-20 text-center">
      <p className="text-5xl font-bold text-gold">404</p>
      <h1 className="mt-3 text-2xl font-bold text-ink">ไม่พบหน้านี้</h1>
      <Link
        to="/overview"
        className="mt-6 flex h-10 items-center rounded-control bg-primary px-4 text-sm font-semibold text-white transition hover:bg-[#4d1621]"
      >
        กลับหน้าภาพรวม
      </Link>
    </div>
  )
}

export default NotFoundPage
