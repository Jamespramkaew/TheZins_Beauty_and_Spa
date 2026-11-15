import { useState } from 'react'
import { FiEye, FiEyeOff, FiLock } from 'react-icons/fi'
import { inputClassName } from './inputClassName'

// Password field with a show/hide toggle. Spread react-hook-form's register() into it.
function PasswordInput({ ref, invalid = false, ...props }) {
  const [visible, setVisible] = useState(false)

  return (
    <div className="relative">
      <FiLock className="pointer-events-none absolute top-1/2 left-4 size-4.5 -translate-y-1/2 text-subtle" aria-hidden="true" />
      <input
        ref={ref}
        type={visible ? 'text' : 'password'}
        aria-invalid={invalid || undefined}
        className={`${inputClassName(invalid)} pr-12`}
        {...props}
      />
      <button
        type="button"
        onClick={() => setVisible((value) => !value)}
        aria-label={visible ? 'ซ่อนรหัสผ่าน' : 'แสดงรหัสผ่าน'}
        className="absolute top-1/2 right-2 flex size-9 -translate-y-1/2 items-center justify-center rounded-lg text-muted transition hover:bg-surface-2 hover:text-ink focus-visible:outline-2 focus-visible:outline-gold-dark disabled:cursor-not-allowed"
      >
        {visible ? <FiEyeOff className="size-4.5" /> : <FiEye className="size-4.5" />}
      </button>
    </div>
  )
}

export default PasswordInput
