import { useSearchParams } from 'react-router-dom'
import Candidates from '../matching/Candidates'
import Requests  from '../matching/Requests'
import Partners  from '../partners/Partners'

const TABS = [
  {
    key: 'find',
    label: 'Find',
    icon: (
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="11" cy="11" r="8"/>
        <path d="m21 21-4.3-4.3"/>
      </svg>
    ),
  },
  {
    key: 'requests',
    label: 'Requests',
    icon: (
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <polyline points="22 12 16 12 14 15 10 15 8 12 2 12"/>
        <path d="M5.45 5.11 2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z"/>
      </svg>
    ),
  },
  {
    key: 'partners',
    label: 'Partners',
    icon: (
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M19 14c1.49-1.46 3-3.21 3-5.5A5.5 5.5 0 0 0 16.5 3c-1.76 0-3 .5-4.5 2-1.5-1.5-2.74-2-4.5-2A5.5 5.5 0 0 0 2 8.5c0 2.3 1.5 4.05 3 5.5l7 7Z"/>
      </svg>
    ),
  },
]

export default function Match() {
  const [params, setParams] = useSearchParams()
  const tab = params.get('tab') || 'find'

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

      {tab === 'find'     && <Candidates />}
      {tab === 'requests' && <Requests />}
      {tab === 'partners' && <Partners />}
    </div>
  )
}
