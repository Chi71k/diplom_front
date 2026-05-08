const PALETTE = [
  '#3b82f6', // blue
  '#8b5cf6', // violet
  '#10b981', // emerald
  '#f59e0b', // amber
  '#ef4444', // red
  '#ec4899', // pink
  '#14b8a6', // teal
  '#6366f1', // indigo
  '#f97316', // orange
  '#06b6d4', // cyan
]

export function avatarColor(str = '') {
  const sum = [...String(str || '?')].reduce((a, c) => a + c.charCodeAt(0), 0)
  return PALETTE[sum % PALETTE.length]
}
