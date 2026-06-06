import { useSearchParams } from 'react-router-dom'
import Groups       from '../groups/Groups'
import Sessions     from '../sessions/Sessions'
import Availability from '../availability/Availability'

const TABS = [
  {
    key: 'groups',
    label: 'Groups',
    icon: (
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M4 15s1-1 4-1 5 2 8 2 4-1 4-1V3s-1 1-4 1-5-2-8-2-4 1-4 1z"/>
        <line x1="4" x2="4" y1="22" y2="15"/>
      </svg>
    ),
  },
  {
    key: 'sessions',
    label: 'Sessions',
    icon: (
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect width="18" height="18" x="3" y="4" rx="2" ry="2"/>
        <line x1="16" x2="16" y1="2" y2="6"/>
        <line x1="8"  x2="8"  y1="2" y2="6"/>
        <line x1="3"  x2="21" y1="10" y2="10"/>
      </svg>
    ),
  },
  {
    key: 'schedule',
    label: 'Schedule',
    icon: (
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10"/>
        <polyline points="12 6 12 12 16 14"/>
      </svg>
    ),
  },
]

export default function Study() {
  const [params, setParams] = useSearchParams()
  const tab = params.get('tab') || 'groups'

  const setTab = (key) => setParams({ tab: key }, { replace: true })

  return (
    <div style={{ maxWidth: 860, margin: '0 auto' }}>
      <div className="page-tab-bar">
        {TABS.map(({ key, label, icon }) => (
          <button
            key={key}
            type="button"
            className={`page-tab-btn${tab === key ? ' active' : ''}`}
            onClick={() => setTab(key)}
          >
            {icon}
            <span>{label}</span>
          </button>
        ))}
      </div>

      {tab === 'groups'   && <Groups />}
      {tab === 'sessions' && <Sessions />}
      {tab === 'schedule' && <Availability />}
    </div>
  )
}
