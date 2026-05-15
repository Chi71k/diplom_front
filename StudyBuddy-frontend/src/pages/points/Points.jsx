import { useState, useEffect } from 'react'
import { useToast } from '../../context/ToastContext'
import { useAuth } from '../../context/useAuth'
import {
  apiGetMyPoints,
  apiGetLeaderboard,
  apiGetReputationLeaderboard,
  apiSearchLeaderboard,
} from '../../api'
import { avatarColor } from '../../utils/avatar'

const reasonMeta = {
  match_accepted:    { label: 'Match accepted',    icon: '🤝' },
  session_completed: { label: 'Session completed', icon: '📅' },
  review_received:   { label: 'Review received',   icon: '⭐' },
  group_created:     { label: 'Group created',     icon: '👥' },
  group_activity:    { label: 'Group activity',    icon: '👥' },
}

const rankMedal = (i) => {
  if (i === 0) return '🥇'
  if (i === 1) return '🥈'
  if (i === 2) return '🥉'
  return null
}

const Points = () => {
  const toast = useToast()
  const { profile } = useAuth()
  const [tab, setTab]               = useState('mine')
  const [myPoints, setMyPoints]     = useState(null)
  const [leaderboard, setLeaderboard] = useState([])
  const [repBoard, setRepBoard]     = useState([])
  const [search, setSearch]         = useState('')
  const [searchResults, setSearchResults] = useState(null)
  const [loading, setLoading]       = useState(true)

  useEffect(() => {
    const load = async () => {
      setLoading(true)
      try {
        const [mine, board, rep] = await Promise.all([
          apiGetMyPoints(),
          apiGetLeaderboard(50),
          apiGetReputationLeaderboard(50),
        ])
        setMyPoints(mine)
        setLeaderboard(board ?? [])
        setRepBoard(rep ?? [])
      } catch (e) {
        toast.error(e.error || 'Failed to load points')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [])

  const handleSearch = async (e) => {
    e.preventDefault()
    if (!search.trim()) { setSearchResults(null); return }
    try {
      const res = await apiSearchLeaderboard(search.trim(), 10)
      setSearchResults(res ?? [])
    } catch (e) {
      toast.error(e.error || 'Search failed')
    }
  }

  const myRank = leaderboard.findIndex((r) => r.userId === profile?.id)

  const fmt = (iso) =>
    new Date(iso).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })

  const displayList = searchResults ?? (tab === 'leaderboard' ? leaderboard : repBoard)

  return (
    <div className="requests-page">
      <div className="card">

        {/* Header */}
        <div className="requests-head">
          <div className="requests-title">Points &amp; Leaderboard</div>
          <div className="requests-sub">Track your activity score and see how you rank</div>
        </div>

        {/* Tabs */}
        <div className="tabs" style={{ marginBottom: '4px' }}>
          {[
            { key: 'mine',        label: 'My Points' },
            { key: 'leaderboard', label: 'Points Board' },
            { key: 'reputation',  label: 'Reputation' },
          ].map(({ key, label }) => (
            <button
              key={key}
              type="button"
              className={`btn btn-sm ${tab === key ? 'btn-primary' : 'btn-secondary'}`}
              onClick={() => setTab(key)}
            >
              {label}
            </button>
          ))}
        </div>

        {loading && <div className="loading-state">Loading…</div>}

        {/* ── My Points ── */}
        {!loading && tab === 'mine' && myPoints && (
          <div>
            {/* Score hero */}
            <div style={{
              margin: '16px 20px',
              padding: '24px',
              borderRadius: '12px',
              background: 'linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%)',
              border: '1px solid #bfdbfe',
              textAlign: 'center',
            }}>
              <div style={{ fontSize: '56px', fontWeight: 800, color: 'var(--primary)', lineHeight: 1 }}>
                {myPoints.totalPoints ?? 0}
              </div>
              <div style={{ color: 'var(--muted)', fontSize: '14px', marginTop: '6px' }}>
                total points
                {myRank >= 0 && (
                  <span style={{ marginLeft: '8px', fontWeight: 700, color: '#d97706' }}>
                    {rankMedal(myRank) ?? `· rank #${myRank + 1}`}
                  </span>
                )}
              </div>
            </div>

            {/* Transactions */}
            <div style={{ padding: '0 20px 16px' }}>
              <div style={{ fontSize: '13px', fontWeight: 700, color: 'var(--muted)', marginBottom: '10px', textTransform: 'uppercase', letterSpacing: '0.5px' }}>
                Recent activity
              </div>

              {(!myPoints.transactions || myPoints.transactions.length === 0) && (
                <div className="empty-state" style={{ padding: '20px 0' }}>No transactions yet.</div>
              )}

              {myPoints.transactions?.map((tx) => {
                const meta = reasonMeta[tx.reason] ?? { label: tx.reason, icon: '💡' }
                return (
                  <div
                    key={tx.id}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '12px',
                      padding: '11px 0',
                      borderBottom: '1px solid var(--border)',
                    }}
                  >
                    <div style={{
                      width: '36px', height: '36px', borderRadius: '10px',
                      background: '#f8fafc', border: '1px solid var(--border)',
                      display: 'flex', alignItems: 'center', justifyContent: 'center',
                      fontSize: '18px', flexShrink: 0,
                    }}>
                      {meta.icon}
                    </div>
                    <div style={{ flex: 1 }}>
                      <div style={{ fontSize: '14px', fontWeight: 500, color: 'var(--text)' }}>
                        {meta.label}
                      </div>
                      <div style={{ fontSize: '12px', color: 'var(--muted)' }}>{fmt(tx.createdAt)}</div>
                    </div>
                    <div style={{
                      fontSize: '15px', fontWeight: 800,
                      color: tx.amount >= 0 ? '#15803d' : '#dc2626',
                    }}>
                      {tx.amount > 0 ? '+' : ''}{tx.amount}
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        )}

        {/* ── Leaderboard / Reputation ── */}
        {!loading && (tab === 'leaderboard' || tab === 'reputation') && (
          <div style={{ padding: '12px 20px 16px' }}>

            {/* Search */}
            <form onSubmit={handleSearch} style={{ display: 'flex', gap: '8px', marginBottom: '16px' }}>
              <input
                style={{
                  flex: 1, padding: '9px 12px',
                  border: '1px solid var(--border)', borderRadius: '8px',
                  fontSize: '14px', outline: 'none',
                }}
                placeholder="Search by name…"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
              <button type="submit" className="btn btn-secondary btn-sm">Search</button>
              {searchResults !== null && (
                <button
                  type="button"
                  className="btn btn-secondary btn-sm"
                  onClick={() => { setSearch(''); setSearchResults(null) }}
                >
                  Clear
                </button>
              )}
            </form>

            {displayList.length === 0 && (
              <div className="empty-state" style={{ padding: '20px 0' }}>No results.</div>
            )}

            {displayList.map((row, i) => {
              const isMe   = row.userId === profile?.id
              const name   = `${row.firstName ?? ''} ${row.lastName ?? ''}`.trim() || row.userId?.slice(0, 8) || '—'
              const medal  = rankMedal(i)
              const score  = tab === 'leaderboard'
                ? `${row.totalPoints} pts`
                : `${row.averageRating?.toFixed(1) ?? '—'} ★`

              return (
                <div
                  key={row.userId ?? i}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '12px',
                    padding: '10px 0',
                    borderBottom: '1px solid var(--border)',
                    background: isMe ? '#eff6ff' : 'transparent',
                    margin: isMe ? '0 -4px' : '0',
                    padding: isMe ? '10px 4px' : '10px 0',
                    borderRadius: isMe ? '6px' : '0',
                  }}
                >
                  {/* Rank */}
                  <div style={{ minWidth: '32px', textAlign: 'center', flexShrink: 0 }}>
                    {medal
                      ? <span style={{ fontSize: '20px' }}>{medal}</span>
                      : <span style={{ fontSize: '14px', fontWeight: 700, color: 'var(--muted)' }}>#{i + 1}</span>
                    }
                  </div>

                  {/* Avatar */}
                  <div
                    className="avatar"
                    style={{ width: '36px', height: '36px', fontSize: '14px', background: avatarColor(row.firstName || row.userId), flexShrink: 0 }}
                  >
                    {name?.[0]?.toUpperCase() ?? '?'}
                  </div>

                  {/* Name */}
                  <div style={{ flex: 1, fontSize: '14px', fontWeight: isMe ? 700 : 400, color: 'var(--text)' }}>
                    {name}
                    {isMe && (
                      <span style={{ marginLeft: '6px', fontSize: '11px', color: 'var(--primary)', fontWeight: 700 }}>
                        you
                      </span>
                    )}
                  </div>

                  {/* Score */}
                  <div style={{ fontSize: '15px', fontWeight: 700, color: 'var(--primary)', flexShrink: 0 }}>
                    {score}
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}

export default Points
