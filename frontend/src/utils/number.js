// Money comes from the API as decimal strings ("4900", "1250.50")
export const toNumber = (value) => {
  const number = Number(value)
  return Number.isFinite(number) ? number : 0
}
