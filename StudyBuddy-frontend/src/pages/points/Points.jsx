import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useToast } from '../../context/ToastContext'
import { useAuth } from '../../context/useAuth'
import {
  apiGetMyPoints,
  apiGetLeaderboard,
  apiGetReputationLeaderboard,
  apiGetUserById,
} from '../../api'
import { avatarColor } from '../../utils/avatar'

const reasonMeta = {
  match_accepted:    { label: 'New study partner accepted' },
  session_confirmed: { label: 'Study session completed' },
  review_left:       { label: 'Left a review' },
  review_received:   { label: 'Received a review' },
  group_created:     { label: 'Created a group' },
  group_activity:    { label: 'Active in a group' },
}

const Points = () => {
  const toast = useToast()
  const { profile } = useAuth()
  const [tab, setTab]               = useState('mine')
  const [myPoints, setMyPoints]     = useState(null)
  const [leaderboard, setLeaderboard] = useState([])
  const [repBoard, setRepBoard]     = useState([])
  const [filter, setFilter]         = useState('')
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
        const allIds = [...new Set([
          ...(board ?? []).map(r => r.userId),
          ...(rep ?? []).map(r => r.userId),
        ])]
        const profileResults = await Promise.allSettled(allIds.map(id => apiGetUserById(id)))
        const profileMap = {}
        allIds.forEach((id, i) => {
          if (profileResults[i].status === 'fulfilled') profileMap[id] = profileResults[i].value
        })
        const enrich = (rows) => (rows ?? []).map(r => ({
          ...r,
          firstName: profileMap[r.userId]?.firstName ?? '',
          lastName:  profileMap[r.userId]?.lastName  ?? '',
          avatarUrl: profileMap[r.userId]?.avatarUrl,
        }))
        setMyPoints(mine)
        setLeaderboard(enrich(board))
        setRepBoard(enrich(rep))
      } catch (e) {
        toast.error(e.error || 'Failed to load points')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [])

  const myRank = leaderboard.findIndex((r) => r.userId === profile?.id)

  const fmt = (iso) =>
    new Date(iso).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })

  const baseList = tab === 'leaderboard' ? leaderboard : repBoard
  const displayList = filter.trim()
    ? baseList.filter((r) =>
        `${r.firstName} ${r.lastName}`.toLowerCase().includes(filter.toLowerCase())
      )
    : baseList

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

        {!loading && tab === 'mine' && myPoints && (
          <div>
            {/* Score hero */}
            <div className="points-hero" style={{
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
                    · rank #{myRank + 1}
                  </span>
                )}
              </div>
            </div>

            <div style={{ padding: '0 20px 16px' }}>
              <div style={{ fontSize: '13px', fontWeight: 700, color: 'var(--muted)', marginBottom: '10px', textTransform: 'uppercase', letterSpacing: '0.5px' }}>
                Recent activity
              </div>

              {(!myPoints.transactions || myPoints.transactions.length === 0) && (
                <div className="empty-state" style={{ padding: '20px 0' }}>No transactions yet.</div>
              )}

              {myPoints.transactions?.map((tx) => {
                const meta = reasonMeta[tx.reason] ?? { label: tx.reason }
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

        {!loading && (tab === 'leaderboard' || tab === 'reputation') && (
          <div style={{ padding: '12px 20px 16px' }}>

            <input
              style={{
                width: '100%', padding: '9px 12px', marginBottom: '16px',
                border: '1px solid var(--border)', borderRadius: '8px',
                fontSize: '14px', outline: 'none', background: 'var(--bg)', color: 'var(--text)',
              }}
              placeholder="Filter by name…"
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
            />

            {displayList.length === 0 && (
              <div className="empty-state" style={{ padding: '20px 0' }}>No results.</div>
            )}

            {displayList.map((row, i) => {
              const isMe  = row.userId === profile?.id
              const name  = `${row.firstName ?? ''} ${row.lastName ?? ''}`.trim() || row.userId?.slice(0, 8) || '—'
              const score = tab === 'leaderboard'
                ? `${row.totalPoints} pts`
                : `${row.averageRating?.toFixed(1) ?? '—'} ★`

              const rowContent = (
                <>
                  <div style={{ minWidth: '32px', textAlign: 'center', flexShrink: 0 }}>
                    <span style={{ fontSize: '14px', fontWeight: 700, color: 'var(--muted)' }}>#{i + 1}</span>
                  </div>
                  <div
                    className="avatar"
                    style={{ width: '36px', height: '36px', fontSize: '14px', background: row.avatarUrl ? undefined : avatarColor(row.firstName || row.userId), flexShrink: 0 }}
                  >
                    {row.avatarUrl ? <img src={row.avatarUrl} alt="" /> : name?.[0]?.toUpperCase() ?? '?'}
                  </div>
                  <div style={{ flex: 1, fontSize: '14px', fontWeight: isMe ? 700 : 400, color: 'var(--text)' }}>
                    {name}
                    {isMe && <span style={{ marginLeft: '6px', fontSize: '11px', color: 'var(--primary)', fontWeight: 700 }}>you</span>}
                  </div>
                  <div style={{ fontSize: '15px', fontWeight: 700, color: 'var(--primary)', flexShrink: 0 }}>{score}</div>
                </>
              )

              const rowStyle = {
                display: 'flex', alignItems: 'center', gap: '12px',
                padding: isMe ? '10px 4px' : '10px 0',
                borderBottom: '1px solid var(--border)',
                background: isMe ? 'var(--blue-50)' : 'transparent',
                margin: isMe ? '0 -4px' : '0',
                borderRadius: isMe ? '6px' : '0',
                textDecoration: 'none',
              }

              return isMe
                ? <div key={row.userId ?? i} style={rowStyle}>{rowContent}</div>
                : (
                  <Link key={row.userId ?? i} to={`/users/${row.userId}`} style={{ ...rowStyle, cursor: 'pointer' }}
                    onMouseEnter={e => e.currentTarget.style.background = 'var(--bg)'}
                    onMouseLeave={e => e.currentTarget.style.background = 'transparent'}
                  >
                    {rowContent}
                  </Link>
                )
            })}
          </div>
        )}
      </div>
    </div>
  )
}

export default Points
