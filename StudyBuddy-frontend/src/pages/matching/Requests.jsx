import { useState, useEffect } from 'react'
import { useToast } from '../../context/ToastContext'
import { apiGetMatchRequests, apiRespondMatchRequest, apiCancelMatchRequest, apiGetUserById } from '../../api'
import { useAuth } from '../../context/useAuth'
import { avatarColor } from '../../utils/avatar'

const statusColor = { pending: '#d97706', accepted: '#15803d', declined: '#dc2626', canceled: '#94a3b8' }
const statusLabel = { pending: 'Pending', accepted: 'Accepted', declined: 'Declined', canceled: 'Canceled' }

const Requests = () => {
  const toast = useToast()
  const { profile } = useAuth()
  const [tab, setTab] = useState('incoming')
  const [requests, setRequests] = useState([])
  const [userCache, setUserCache] = useState({})
  const [loading, setLoading] = useState(true)
  const [acting, setActing] = useState(null)

  const load = async () => {
    setLoading(true)
    try {
      const data = await apiGetMatchRequests({ limit: 50 })
      const items = data.items ?? []
      setRequests(items)
      const ids = [...new Set(items.flatMap((r) => [r.requesterId, r.receiverId]))]
      const entries = await Promise.allSettled(ids.map((id) => apiGetUserById(id).then((u) => [id, u])))
      const cache = {}
      for (const r of entries) {
        if (r.status === 'fulfilled') {
          const [id, user] = r.value
          cache[id] = user
        }
      }
      setUserCache(cache)
    } catch (e) {
      toast.error(e.error || 'Failed to load requests')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const myId = profile?.id
  const incoming = requests.filter((r) => r.receiverId === myId)
  const outgoing = requests.filter((r) => r.requesterId === myId)
  const shown = tab === 'incoming' ? incoming : outgoing

  const handleRespond = async (id, accept) => {
    setActing(id)
    try {
      await apiRespondMatchRequest(id, accept)
      toast.success(accept ? 'Request accepted!' : 'Request declined')
      setRequests((prev) => prev.map((r) => r.id === id ? { ...r, status: accept ? 'accepted' : 'declined' } : r))
    } catch (e) {
      toast.error(e.error || 'Failed to respond')
    } finally {
      setActing(null)
    }
  }

  const handleCancel = async (id) => {
    setActing(id)
    try {
      await apiCancelMatchRequest(id)
      toast.success('Request canceled')
      setRequests((prev) => prev.map((r) => r.id === id ? { ...r, status: 'canceled' } : r))
    } catch (e) {
      toast.error(e.error || 'Failed to cancel')
    } finally {
      setActing(null)
    }
  }

  const fmt = (iso) =>
    new Date(iso).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })

  return (
    <div className="requests-page">
      <div className="card">
        <div className="requests-head">
          <div className="requests-title">Match Requests</div>
          <div className="requests-sub">Manage your incoming and outgoing study partner requests</div>
        </div>

        <div className="tabs">
          {['incoming', 'outgoing'].map((t) => (
            <button
              key={t}
              type="button"
              className={`btn ${tab === t ? 'btn-primary' : 'btn-secondary'} btn-sm`}
              onClick={() => setTab(t)}
            >
              {t === 'incoming'
                ? `Incoming (${incoming.length})`
                : `Outgoing (${outgoing.length})`
              }
            </button>
          ))}
        </div>

        {loading && <div className="loading-state">Loading requests...</div>}

        {!loading && shown.length === 0 && (
          <div className="empty-state">No {tab} requests yet.</div>
        )}

        {!loading && shown.map((r) => {
          const otherId = tab === 'incoming' ? r.requesterId : r.receiverId
          const other = userCache[otherId]
          return (
            <div key={r.id} className="req-card">
              <div className="req-card-main">
                <div className="avatar avatar-md" style={{ background: avatarColor(other?.firstName || otherId) }}>
                  {other?.avatarUrl
                    ? <img src={other.avatarUrl} alt="" />
                    : (other?.firstName?.[0] || '?').toUpperCase()
                  }
                </div>

                <div className="req-card-info">
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
                    <span className="req-card-name">
                      {other ? `${other.firstName} ${other.lastName}` : otherId}
                    </span>
                    <span style={{
                      fontSize: '11px', fontWeight: 600, padding: '2px 7px',
                      borderRadius: '5px', background: '#f8fafc', border: '1px solid var(--border)',
                      color: statusColor[r.status] || '#94a3b8',
                    }}>
                      {statusLabel[r.status] || r.status}
                    </span>
                  </div>
                  <div className="req-card-role">{fmt(r.createdAt)}</div>
                  {r.message && <div className="req-card-msg">"{r.message}"</div>}

                  <div className="req-card-actions">
                    {r.status === 'pending' && tab === 'incoming' && (
                      <>
                        <button
                          type="button"
                          className="btn btn-primary btn-sm"
                          disabled={acting === r.id}
                          onClick={() => handleRespond(r.id, true)}
                        >Accept</button>
                        <button
                          type="button"
                          className="btn btn-danger btn-sm"
                          disabled={acting === r.id}
                          onClick={() => handleRespond(r.id, false)}
                        >Decline</button>
                      </>
                    )}
                    {r.status === 'pending' && tab === 'outgoing' && (
                      <button
                        type="button"
                        className="btn btn-secondary btn-sm"
                        disabled={acting === r.id}
                        onClick={() => handleCancel(r.id)}
                      >Cancel</button>
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

export default Requests
