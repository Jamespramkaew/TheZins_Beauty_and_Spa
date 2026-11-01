import { useContext } from 'react'
import { ToastContext } from '../components/toast/ToastContext'

// Usage: const toast = useToast()
//        toast.success('บันทึกสำเร็จ', 'เพิ่มคอร์สใหม่แล้ว')
//        toast.error('บันทึกไม่สำเร็จ', error.message)
export function useToast() {
  const context = useContext(ToastContext)
  if (!context) throw new Error('useToast must be used within ToastProvider')
  return context
}
