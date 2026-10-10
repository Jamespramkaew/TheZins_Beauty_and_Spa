import { Link } from 'react-router-dom'

function NotFoundPage() {
  return (
    <div className="text-center">
      <h1 className="text-2xl font-bold">404 - ไม่พบหน้านี้</h1>
      <Link to="/" className="text-pink-500 underline">
        กลับหน้าแรก
      </Link>
    </div>
  )
}

export default NotFoundPage
