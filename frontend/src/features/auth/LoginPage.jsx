import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Navigate, useLocation, useNavigate } from 'react-router-dom'
import { FiAlertCircle, FiArrowRight, FiCheck, FiClock, FiHome, FiLoader, FiUser } from 'react-icons/fi'
import FullPageLoader from '../../components/FullPageLoader'
import PasswordInput from '../../components/ui/PasswordInput'
import { inputClassName } from '../../components/ui/inputClassName'
import { useAuth } from '../../hooks/useAuth'

const LOGIN_FAILED_MESSAGE = 'ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง กรุณาลองใหม่อีกครั้ง'
const NETWORK_ERROR_MESSAGE = 'เชื่อมต่อเซิร์ฟเวอร์ไม่ได้ กรุณาลองใหม่อีกครั้ง'
const SERVER_ERROR_MESSAGE = 'เกิดข้อผิดพลาดในระบบ กรุณาลองใหม่อีกครั้ง'
// How long the success state shows before moving to Overview
const SUCCESS_DELAY_MS = 700
const focusRing = 'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-gold-dark'

// 401 = wrong username/password; anything else is a connection or server problem
const toLoginError = (error) => {
  if (error.response?.status === 401) return { type: 'credentials', message: LOGIN_FAILED_MESSAGE }
  if (!error.response) return { type: 'network', message: NETWORK_ERROR_MESSAGE }
  return { type: 'server', message: SERVER_ERROR_MESSAGE }
}

function LoginPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const { status, signIn } = useAuth()
  const [succeeded, setSucceeded] = useState(false)
  // Where RequireAuth sent the user from; go back there after logging in
  const redirectTo = location.state?.from?.pathname ?? '/overview'
  const sessionExpired = Boolean(location.state?.expired)

  const {
    register,
    handleSubmit,
    setError,
    clearErrors,
    formState: { errors, isSubmitting },
  } = useForm({
    mode: 'onSubmit',
    reValidateMode: 'onChange',
    defaultValues: { username: '', password: '' },
  })

  const onSubmit = async ({ username, password }) => {
    try {
      await signIn({ username, password })
    } catch (error) {
      setError('root', toLoginError(error))
      return
    }

    // Keep the form locked (isSubmitting) while the success state shows, then fade into the app
    setSucceeded(true)
    await new Promise((resolve) => setTimeout(resolve, SUCCESS_DELAY_MS))
    navigate(redirectTo, { replace: true, viewTransition: true })
  }

  // Hide the login error as soon as the user edits their credentials
  const clearLoginError = () => clearErrors('root')

  const usernameInvalid = Boolean(errors.username)
  const passwordInvalid = Boolean(errors.password || errors.root?.type === 'credentials')
  const passwordDescribedBy = [errors.password && 'password-error', passwordInvalid && errors.root && 'login-error']
    .filter(Boolean)
    .join(' ')

  if (status === 'checking') return <FullPageLoader />
  // Already logged in (e.g. opened "/" with a valid session). While submitting, the
  // success animation handles navigation itself.
  if (status === 'authenticated' && !isSubmitting) return <Navigate to={redirectTo} replace />

  return (
    <div className="flex min-h-screen flex-wrap bg-white">
      {/* Below 960px the panels stack. The banner is square with the logo in its
          middle third, so pin the logo near the top and size the panel so the
          caption never covers it. */}
      <section className="relative isolate flex min-h-70 flex-[1_1_520px] overflow-hidden bg-sidebar max-[960px]:min-h-[max(280px,calc(36vw+150px))]">
        <img
          src="/Logo-Full-Banner.jpg"
          alt="โลโก้ THE ZIN Beauty & Spa"
          className="absolute inset-0 -z-10 size-full object-cover object-center max-[960px]:object-[50%_calc(24px-32vw)]"
        />
        <div className="mt-auto grid w-full grid-cols-[1fr_auto] items-end gap-x-6 gap-y-1.5 bg-linear-to-t from-[rgba(26,6,9,0.75)] to-transparent px-6 pt-20 pb-7 sm:px-9 sm:pb-8">
          <p className="col-start-1 row-start-1 text-xs font-semibold tracking-[0.16em] text-gold">BACK OFFICE</p>
          <p className="col-span-2 row-start-2 text-[15px] text-[#F4EAE0] min-[960px]:col-span-1">
            ระบบจัดการหลังบ้าน · ออเดอร์ สต็อก คอร์ส และสาขา
          </p>
          <p className="col-start-2 row-start-1 text-xs text-[#D9C7BE] min-[960px]:row-start-2 min-[960px]:pb-0.5">
            © 2569 THE ZIN Beauty &amp; Spa
          </p>
        </div>
      </section>

      <main className="flex flex-[1_1_440px] items-center justify-center px-6 py-12">
        <div className="w-full max-w-100 space-y-6">
          <p className="flex items-center gap-2 text-xs font-bold tracking-[0.16em] text-gold-dark">
            <span className="h-px w-4.5 bg-gold" aria-hidden="true" />
            THE ZIN
          </p>

          <div>
            <h1 className="text-[30px] leading-tight font-bold text-ink">เข้าสู่ระบบ</h1>
            <p className="mt-2 text-sm text-muted">ใช้บัญชีผู้ดูแล หรือบัญชีของสาขาคุณ</p>
          </div>

          <form onSubmit={handleSubmit(onSubmit)} noValidate>
            <fieldset disabled={isSubmitting} className="min-w-0 space-y-6">
              {sessionExpired && !errors.root && (
                <div className="flex items-start gap-2.5 rounded-control bg-gold-bg px-4 py-3 text-sm text-gold-ink">
                  <FiClock className="mt-0.5 size-4 shrink-0" aria-hidden="true" />
                  <p>เซสชันหมดอายุ กรุณาเข้าสู่ระบบอีกครั้ง</p>
                </div>
              )}

              {errors.root && (
                <div
                  id="login-error"
                  role="alert"
                  className="flex items-start gap-2.5 rounded-control border border-error-border bg-danger-bg px-4 py-3 text-sm text-error"
                >
                  <FiAlertCircle className="mt-0.5 size-4 shrink-0" aria-hidden="true" />
                  <p>{errors.root.message}</p>
                </div>
              )}

              <div>
                <label htmlFor="username" className="mb-2 block text-[13px] font-semibold text-ink">
                  ชื่อผู้ใช้
                </label>
                <div className="relative">
                  <FiUser
                    className="pointer-events-none absolute top-1/2 left-4 size-4.5 -translate-y-1/2 text-subtle"
                    aria-hidden="true"
                  />
                  <input
                    id="username"
                    type="text"
                    placeholder="เช่น admin"
                    autoComplete="username"
                    aria-invalid={usernameInvalid || undefined}
                    aria-describedby={usernameInvalid ? 'username-error' : undefined}
                    className={`${inputClassName(usernameInvalid)} pr-4`}
                    {...register('username', {
                      setValueAs: (value) => value.trim(),
                      required: 'กรุณากรอกชื่อผู้ใช้',
                      maxLength: { value: 50, message: 'ชื่อผู้ใช้ต้องไม่เกิน 50 ตัวอักษร' },
                      onChange: clearLoginError,
                    })}
                  />
                </div>
                {errors.username && (
                  <p id="username-error" className="mt-1.5 text-[13px] text-error">
                    {errors.username.message}
                  </p>
                )}
              </div>

              <div>
                <label htmlFor="password" className="mb-2 block text-[13px] font-semibold text-ink">
                  รหัสผ่าน
                </label>
                <PasswordInput
                  id="password"
                  placeholder="รหัสผ่าน"
                  autoComplete="current-password"
                  invalid={passwordInvalid}
                  aria-describedby={passwordDescribedBy || undefined}
                  {...register('password', {
                    required: 'กรุณากรอกรหัสผ่าน',
                    onChange: clearLoginError,
                  })}
                />
                {errors.password && (
                  <p id="password-error" className="mt-1.5 text-[13px] text-error">
                    {errors.password.message}
                  </p>
                )}
              </div>

              <button
                type="submit"
                disabled={isSubmitting}
                className={`flex h-12 w-full items-center justify-center gap-2 rounded-control text-[15px] font-semibold text-white transition-colors duration-300 disabled:cursor-not-allowed ${
                  succeeded ? 'bg-success' : 'bg-primary hover:bg-[#4d1621] disabled:opacity-85'
                } ${focusRing}`}
              >
                {succeeded ? (
                  <>
                    <FiCheck className="size-4.5 motion-safe:animate-check-pop" aria-hidden="true" />
                    เข้าสู่ระบบสำเร็จ
                  </>
                ) : isSubmitting ? (
                  <>
                    <FiLoader className="size-4 animate-spin" aria-hidden="true" />
                    กำลังเข้าสู่ระบบ…
                  </>
                ) : (
                  <>
                    เข้าสู่ระบบ
                    <FiArrowRight className="size-4" aria-hidden="true" />
                  </>
                )}
              </button>
              <p role="status" className="sr-only">
                {succeeded ? 'เข้าสู่ระบบสำเร็จ กำลังไปหน้าภาพรวม' : ''}
              </p>
            </fieldset>
          </form>

          <div className="flex items-start gap-2.5 border-t border-divider pt-6 text-[13px] leading-relaxed text-muted">
            <FiHome className="mt-0.5 size-4 shrink-0" aria-hidden="true" />
            <p>บัญชีสาขาใช้ร่วมกันทั้งสาขา — ตอนบันทึกรายการ ระบบจะให้เลือกชื่อพนักงานผู้ทำรายการ</p>
          </div>
        </div>
      </main>
    </div>
  )
}

export default LoginPage
