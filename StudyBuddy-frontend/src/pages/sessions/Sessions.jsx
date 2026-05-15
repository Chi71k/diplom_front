import { useState, useEffect } from 'react'
import { useToast } from '../../context/ToastContext'
import { useAuth } from '../../context/useAuth'
import {
  apiListMySessions,
  apiProposeSession,
  apiConfirmSession,
  apiCancelSession,
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
  const [sessions, setSessions]   = useState([])
  const [loading, setLoading]     = useState(true)
  const [acting, setActing]       = useState(null)
  const [showForm, setShowForm]   = useState(false)
  const [form, setForm] = useState({
    title: '',
    participantIds: '',
    startTime: '',
    endTime: '',
  })

  const load = async () => {
    setLoading(true)
    try {
      const data = await apiListMySessions()
      setSessions(data.items ?? [])
    } catch (e) {
      toast.error(e.error || 'Failed to load sessions')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const handlePropose = async (e) => {
    e.preventDefault()
    if (!form.startTime || !form.endTime) { toast.error('Set start and end time'); return }
    try {
      const tz = Intl.DateTimeFormat().resolvedOptions().timeZone
      const s = await apiProposeSession({
        title: form.title,
        participantIds: form.participantIds.split(',').map((x) => x.trim()).filter(Boolean),
        startTime: new Date(form.startTime).toISOString(),
        endTime: new Date(form.endTime).toISOString(),
        timezone: tz,
      })
      toast.success('Session proposed!')
      setSessions((prev) => [s, ...prev])
      setShowForm(false)
      setForm({ title: '', participantIds: '', startTime: '', endTime: '' })
    } catch (e) {
      toast.error(e.error || 'Failed to propose session')
    }
  }

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

  const fmtRange = (start, end) => {
    const s = new Date(start)
    const e = new Date(end)
    const date = s.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
    const t1 = s.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })
    const t2 = e.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })
    return `${date} · ${t1} – ${t2}`
  }

  return (
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

        {/* Propose form */}
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
                Participant IDs
                <span style={{ color: 'var(--muted-lt)', marginLeft: '4px' }}>(comma-separated, optional)</span>
              </label>
              <input
                style={inputStyle}
                placeholder="uuid1, uuid2, …"
                value={form.participantIds}
                onChange={(e) => setForm((f) => ({ ...f, participantIds: e.target.value }))}
              />
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
                <div style={{
                  width: '40px', height: '40px', borderRadius: '10px',
                  background: '#eff6ff', display: 'flex',
                  alignItems: 'center', justifyContent: 'center',
                  fontSize: '18px', flexShrink: 0,
                }}>
                  📅
                </div>

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
                    🕐 {fmtRange(s.startTime, s.endTime)}
                  </div>

                  <div style={{ fontSize: '12px', color: 'var(--muted)', marginTop: '3px' }}>
                    {count} participant{count !== 1 ? 's' : ''}
                    {s.timezone && ` · ${s.timezone}`}
                  </div>

                  <div className="req-card-actions">
                    {s.status === 'proposed' && !isConfirmed && (
                      <button
                        className="btn btn-primary btn-sm"
                        disabled={acting === s.id}
                        onClick={() => handleConfirm(s.id)}
                      >
                        ✓ Confirm
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
  )
}

export default Sessions
