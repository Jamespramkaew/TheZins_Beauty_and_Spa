const numberFormatter = new Intl.NumberFormat('th-TH')

export const formatNumber = (value) => numberFormatter.format(value)

export const formatBaht = (value) => `${formatNumber(value)}฿`

export const formatPercent = (value, digits = 1) => `${value.toFixed(digits)}%`
