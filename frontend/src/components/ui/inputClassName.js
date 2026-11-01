// Text input with a left icon; shared by the Login fields and PasswordInput
export const inputClassName = (invalid) =>
  `h-12 w-full rounded-control bg-white pl-11 text-[15px] text-ink transition placeholder:text-subtle focus:ring-3 focus:outline-none disabled:cursor-not-allowed disabled:opacity-60 ${
    invalid
      ? 'border-[1.5px] border-error-strong bg-error-input-bg focus:ring-error-strong/15'
      : 'border border-input focus:border-gold-dark focus:ring-gold/30'
  }`
