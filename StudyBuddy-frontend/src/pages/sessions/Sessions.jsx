import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useToast } from '../../context/ToastContext'
import { useAuth } from '../../context/useAuth'
import {
  apiListMySessions,
  apiProposeSession,
  apiConfirmSession,
  apiCancelSession,
  apiExportSessionToGCal,
  apiGetFriends,
  apiGetUserById,
} from '../../api'

const STATUS = {
  proposed:  { label: 'Proposed',  bg: '#fef9c3', color: '#a16207' },
  confirmed: { label: 'Confirmed', bg: '#dcfce7', color: '#15803d' },
  canceled:  { label: 'Canceled',  bg: '#f1f5f9', color: '#94a3b8' },
}

const StatusBadge = ({ status }) => {
  const s = STATUS[status] ?? { label: status, bg: '#f1f5f9', color: '#94a3b8' }
  return (
    <span style={{
      fontSize: '11px', fontWeight: 700, padding: '3px 8px',
      borderRadius: '5px', background: s.bg, color: s.color,
    }}>
      {s.label}
    </span>
  )
}

const inputStyle = {
  width: '100%', padding: '9px 12px',
  border: '1px solid var(--border)', borderRadius: '8px',
  fontSize: '14px', outline: 'none', background: '#fff',
}

const Sessions = () => {
  const toast = useToast()
  const { profile } = useAuth()
  const [sessions, setSessions]         = useState([])
  const [loading, setLoading]           = useState(true)
  const [acting, setActing]             = useState(null)
  const [showForm, setShowForm]         = useState(false)
  const [detailSession, setDetail]      = useState(null)
  const [detailUsers, setDetailUsers]   = useState({})
  const [friends, setFriends]           = useState([])
  const [selectedFriends, setSelectedFriends] = useState([])
  const [form, setForm] = useState({
    title: '',
    startTime: '',
    endTime: '',
  })

  const load = async () => {
    setLoading(true)
    try {
      const [sessData, friendsData] = await Promise.all([
        apiListMySessions(),
        apiGetFriends().catch(() => ({ items: [] })),
      ])
      setSessions(sessData.items ?? [])
      const friendIds = friendsData.items ?? friendsData ?? []
      if (friendIds.length > 0) {
        const results = await Promise.allSettled(friendIds.map(id => apiGetUserById(id)))
        setFriends(friendIds.map((id, i) => {
          const p = results[i].status === 'fulfilled' ? results[i].value : null
          return { id, firstName: p?.firstName ?? '', lastName: p?.lastName ?? '', avatarUrl: p?.avatarUrl }
        }))
      } else {
        setFriends([])
      }
    } catch (e) {
      toast.error(e.error || 'Failed to load sessions')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  useEffect(() => {
    if (!detailSession) return
    const ids = (detailSession.participants ?? [])
      .map(p => p.userId)
      .filter(id => id !== profile?.id)
    if (!ids.length) return
    Promise.allSettled(ids.map(id => apiGetUserById(id))).then(results => {
      const map = {}
      ids.forEach((id, i) => {
        if (results[i].status === 'fulfilled') map[id] = results[i].value
      })
      setDetailUsers(map)
    })
  }, [detailSession?.id])

  const handlePropose = async (e) => {
    e.preventDefault()
    if (!form.startTime || !form.endTime) { toast.error('Set start and end time'); return }
    if (selectedFriends.length === 0) { toast.error('Session must have at least one other participant'); return }
    try {
      const tz = Intl.DateTimeFormat().resolvedOptions().timeZone
      const s = await apiProposeSession({
        title: form.title,
        participantIds: selectedFriends,
        startTime: new Date(form.startTime).toISOString(),
        endTime: new Date(form.endTime).toISOString(),
        timezone: tz,
      })
      toast.success('Session proposed!')
      setSessions((prev) => [s, ...prev])
      setShowForm(false)
      setForm({ title: '', startTime: '', endTime: '' })
      setSelectedFriends([])
    } catch (e) {
      toast.error(e.error || 'Failed to propose session')
    }
  }

  const toggleFriend = (id) =>
    setSelectedFriends(prev => prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id])

  const handleConfirm = async (id) => {
    setActing(id)
    try {
      const s = await apiConfirmSession(id)
      toast.success('Session confirmed!')
      setSessions((prev) => prev.map((x) => (x.id === id ? s : x)))
    } catch (e) {
      toast.error(e.error || 'Failed to confirm')
    } finally {
      setActing(null)
    }
  }

  const handleCancel = async (id) => {
    setActing(id)
    try {
      await apiCancelSession(id)
      toast.success('Session canceled')
      setSessions((prev) => prev.map((x) => (x.id === id ? { ...x, status: 'canceled' } : x)))
    } catch (e) {
      toast.error(e.error || 'Failed to cancel')
    } finally {
      setActing(null)
    }
  }

  const handleExportGCal = async (id) => {
    setActing(id)
    try {
      await apiExportSessionToGCal(id)
      toast.success('Session exported to Google Calendar!')
    } catch (e) {
      toast.error(e.error || 'Export failed — make sure Google Calendar is connected')
    } finally {
      setActing(null)
    }
  }

  const fmtRange = (start, end) => {
    const s = new Date(start)
    const e = new Date(end)
    const date = s.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
    const t1 = s.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })
    const t2 = e.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })
    return `${date} · ${t1} – ${t2}`
  }

  return (
    <>
    <div className="requests-page">
      <div className="card">

        {/* Header */}
        <div className="requests-head" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <div className="requests-title">Study Sessions</div>
            <div className="requests-sub">Propose and manage your study sessions</div>
          </div>
          <button
            className={`btn btn-sm ${showForm ? 'btn-secondary' : 'btn-primary'}`}
            onClick={() => setShowForm((v) => !v)}
          >
            {showForm ? 'Cancel' : '+ New session'}
          </button>
        </div>

        {showForm && (
          <form
            onSubmit={handlePropose}
            style={{
              margin: '0 20px 20px',
              padding: '18px',
              background: '#f8fafc',
              borderRadius: '10px',
              border: '1px solid var(--border)',
              display: 'flex',
              flexDirection: 'column',
              gap: '14px',
            }}
          >
            <div>
              <label style={{ fontSize: '12px', color: 'var(--muted)', marginBottom: '6px', display: 'block' }}>
                Title *
              </label>
              <input
                style={inputStyle}
                placeholder="e.g. Algorithms review session"
                required
                value={form.title}
                onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))}
              />
            </div>

            <div>
              <label style={{ fontSize: '12px', color: 'var(--muted)', marginBottom: '6px', display: 'block' }}>
                Invite partners
                <span style={{ color: 'var(--danger)', marginLeft: '4px' }}>* required</span>
              </label>
              {friends.length === 0
                ? <div style={{ fontSize: '13px', color: 'var(--muted)' }}>No partners yet — connect with someone first.</div>
                : <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', maxHeight: '160px', overflowY: 'auto' }}>
                    {friends.map(f => {
                      const name = `${f.firstName ?? ''} ${f.lastName ?? ''}`.trim() || f.id?.slice(0, 8)
                      const checked = selectedFriends.includes(f.id)
                      return (
                        <label key={f.id} style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer', fontSize: '14px', padding: '4px 6px', borderRadius: '6px', background: checked ? 'var(--blue-50)' : 'transparent' }}>
                          <input type="checkbox" checked={checked} onChange={() => toggleFriend(f.id)} style={{ width: 'auto' }} />
                          <span>{name}</span>
                        </label>
                      )
                    })}
                  </div>
              }
            </div>

            <div style={{ display: 'flex', gap: '12px' }}>
              <div style={{ flex: 1 }}>
                <label style={{ fontSize: '12px', color: 'var(--muted)', marginBottom: '6px', display: 'block' }}>Start *</label>
                <input
                  type="datetime-local"
                  style={inputStyle}
                  required
                  value={form.startTime}
                  onChange={(e) => setForm((f) => ({ ...f, startTime: e.target.value }))}
                />
              </div>
              <div style={{ flex: 1 }}>
                <label style={{ fontSize: '12px', color: 'var(--muted)', marginBottom: '6px', display: 'block' }}>End *</label>
                <input
                  type="datetime-local"
                  style={inputStyle}
                  required
                  value={form.endTime}
                  onChange={(e) => setForm((f) => ({ ...f, endTime: e.target.value }))}
                />
              </div>
            </div>

            <button type="submit" className="btn btn-primary" style={{ alignSelf: 'flex-start' }}>
              Propose session
            </button>
          </form>
        )}

        {loading && <div className="loading-state">Loading sessions…</div>}

        {!loading && sessions.length === 0 && (
          <div className="empty-state">No sessions yet. Propose one to get started!</div>
        )}

        {!loading && sessions.map((s) => {
          const isOrganizer = s.organizerId === profile?.id
          const myPart      = s.participants?.find((p) => p.userId === profile?.id)
          const isConfirmed = myPart?.confirmed
          const count       = s.participants?.length ?? 0

          return (
            <div key={s.id} className="req-card">
              <div className="req-card-main">
                <div className="req-card-info" style={{ width: '100%' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
                    <span className="req-card-name">{s.title}</span>
                    <StatusBadge status={s.status} />
                    {isOrganizer && (
                      <span style={{ fontSize: '11px', color: 'var(--muted)', background: '#f8fafc', padding: '2px 7px', borderRadius: '5px', border: '1px solid var(--border)' }}>
                        organizer
                      </span>
                    )}
                  </div>

                  <div className="req-card-role" style={{ marginTop: '4px' }}>
                    {fmtRange(s.startTime, s.endTime)}
                  </div>

                  <div style={{ fontSize: '12px', color: 'var(--muted)', marginTop: '3px' }}>
                    {count} participant{count !== 1 ? 's' : ''}
                    {s.timezone && ` · ${s.timezone}`}
                  </div>

                  <div className="req-card-actions">
                    <button
                      className="btn btn-secondary btn-sm"
                      onClick={() => setDetail(s)}
                    >
                      Details
                    </button>
                    {s.status === 'proposed' && !isConfirmed && (
                      <button
                        className="btn btn-primary btn-sm"
                        disabled={acting === s.id}
                        onClick={() => handleConfirm(s.id)}
                      >
                        Confirm
                      </button>
                    )}
                    {s.status === 'confirmed' && (
                      <button
                        className="btn btn-secondary btn-sm"
                        disabled={acting === s.id}
                        onClick={() => handleExportGCal(s.id)}
                        title="Export to Google Calendar"
                      >
                        📅 Export to GCal
                      </button>
                    )}
                    {s.status !== 'canceled' && (isOrganizer || s.status === 'proposed') && (
                      <button
                        className="btn btn-danger btn-sm"
                        disabled={acting === s.id}
                        onClick={() => handleCancel(s.id)}
                      >
                        Cancel
                      </button>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )
        })}
      </div>
    </div>

    {detailSession && (
      <div
        style={{
          position: 'fixed', inset: 0, background: 'rgba(0,0,0,.45)',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          zIndex: 1000, padding: '16px',
        }}
        onClick={() => setDetail(null)}
      >
        <div
          style={{
            background: '#fff', borderRadius: '14px', padding: '28px',
            maxWidth: '520px', width: '100%',
            boxShadow: '0 20px 60px rgba(0,0,0,.25)',
            maxHeight: '80vh', overflowY: 'auto',
          }}
          onClick={(e) => e.stopPropagation()}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '20px' }}>
            <div>
              <div style={{ fontSize: '18px', fontWeight: 700, marginBottom: '6px' }}>
                {detailSession.title}
              </div>
              <StatusBadge status={detailSession.status} />
            </div>
            <button
              onClick={() => setDetail(null)}
              style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: '20px', color: 'var(--muted)', lineHeight: 1, padding: '2px 6px' }}
            >
              ✕
            </button>
          </div>

          {[
            ['When',     fmtRange(detailSession.startTime, detailSession.endTime)],
            ['Timezone', detailSession.timezone],
            detailSession.courseId && ['Course ID', detailSession.courseId],
            detailSession.groupId  && ['Group ID',  detailSession.groupId],
          ].filter(Boolean).map(([label, value]) => (
            <div key={label} style={{ display: 'flex', gap: '12px', marginBottom: '12px', fontSize: '14px' }}>
              <span style={{ color: 'var(--muted)', minWidth: '80px', flexShrink: 0 }}>{label}</span>
              <span style={{ fontWeight: 500, wordBreak: 'break-all' }}>{value}</span>
            </div>
          ))}

          <div style={{ marginTop: '16px' }}>
            <div style={{ fontSize: '12px', color: 'var(--muted)', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '.05em', marginBottom: '10px' }}>
              Participants ({detailSession.participants?.length ?? 0})
            </div>
            {detailSession.participants?.map((p) => {
              const u = p.userId === profile?.id ? profile : detailUsers[p.userId]
              const pName = p.userId === profile?.id
                ? 'You'
                : u ? `${u.firstName ?? ''} ${u.lastName ?? ''}`.trim() || `${p.userId.slice(0,8)}…`
                : `${p.userId.slice(0,8)}…`
              return (
                <div key={p.userId} style={{
                  display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                  padding: '8px 0', borderBottom: '1px solid var(--border)', fontSize: '13px',
                }}>
                  <span style={{ color: 'var(--text)' }}>
                    {p.userId !== profile?.id
                      ? <Link to={`/users/${p.userId}`} style={{ color: 'var(--primary)', fontWeight: 500 }}>{pName}</Link>
                      : pName
                    }
                    {p.userId === detailSession.organizerId && (
                      <span style={{ marginLeft: '6px', fontSize: '11px', color: 'var(--muted)', background: 'var(--bg)', padding: '1px 6px', borderRadius: '4px', border: '1px solid var(--border)' }}>
                        organizer
                      </span>
                    )}
                  </span>
                  <span style={{
                    fontWeight: 600, fontSize: '11px',
                    color: p.confirmed ? '#15803d' : '#a16207',
                    background: p.confirmed ? '#dcfce7' : '#fef9c3',
                    padding: '2px 8px', borderRadius: '5px',
                  }}>
                    {p.confirmed ? '✓ Confirmed' : 'Pending'}
                  </span>
                </div>
              )
            })}
          </div>

          <div style={{ marginTop: '16px', fontSize: '11px', color: 'var(--muted)', wordBreak: 'break-all' }}>
            ID: {detailSession.id}
          </div>
        </div>
      </div>
    )}
    </>
  )
}

export default Sessions
