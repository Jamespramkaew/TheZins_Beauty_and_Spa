// Money comes from the API as decimal strings ("4900", "1250.50")
export const toNumber = (value) => {
  const number = Number(value)
  return Number.isFinite(number) ? number : 0
}

// Round to satang (2 decimals). JS floats drift: 3 × 0.10 = 0.30000000000000004,
// which the backend's "at most 2 decimals" rule would reject.
export const roundMoney = (value) => Math.round(value * 100) / 100
