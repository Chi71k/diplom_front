const PALETTE = [
  '#3b82f6',
  '#8b5cf6',
  '#10b981',
  '#f59e0b',
  '#ef4444',
  '#ec4899',
  '#14b8a6',
  '#6366f1',
  '#f97316',
  '#06b6d4',
]

export function avatarColor(str = '') {
  const sum = [...String(str || '?')].reduce((a, c) => a + c.charCodeAt(0), 0)
  return PALETTE[sum % PALETTE.length]
}
