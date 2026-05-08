import { useState, useEffect } from 'react'
import { useToast } from '../../context/ToastContext'
import { apiGetCandidates, apiSendMatchRequest } from '../../api'
import { avatarColor } from '../../utils/avatar'

const matchClass = (score) => {
  if (score >= 0.7) return 'match-green'
  if (score >= 0.4) return 'match-amber'
  return 'match-gray'
}

const Candidates = () => {
  const toast = useToast()
  const [candidates, setCandidates] = useState([])
  const [loading, setLoading] = useState(true)
  const [sending, setSending] = useState(null)
  const [messages, setMessages] = useState({})
  const [openMsg, setOpenMsg] = useState(null)

  useEffect(() => {
    const load = async () => {
      setLoading(true)
      try {
        const data = await apiGetCandidates(20)
        setCandidates(data.items ?? [])
      } catch (e) {
        toast.error(e.error || 'Failed to load candidates')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [])

  const handleSend = async (userId) => {
    setSending(userId)
    try {
      await apiSendMatchRequest(userId, messages[userId] || '')
      toast.success('Request sent!')
      setCandidates((prev) => prev.filter((c) => c.userId !== userId))
      setOpenMsg(null)
    } catch (e) {
      toast.error(e.error || 'Failed to send request')
    } finally {
      setSending(null)
    }
  }

  return (
    <div className="find-page">
      <div className="card">
        <div className="find-header">
          <div>
            <div className="find-title">Find study partners</div>
            <div className="find-sub">Ranked by shared interests (40%), availability (40%), and courses (20%)</div>
          </div>
          <span style={{ fontSize: '12px', color: 'var(--muted)' }}>
            {!loading && `${candidates.length} candidate${candidates.length !== 1 ? 's' : ''}`}
          </span>
        </div>

        {loading && <div className="loading-state">Finding candidates...</div>}

        {!loading && candidates.length === 0 && (
          <div className="empty-state">
            No candidates found. Make sure your interests, courses, and availability are filled in.
          </div>
        )}

        {!loading && candidates.map((c) => (
          <div key={c.userId} className="cand-card">
            <div className="cand-main">
              <div className="avatar avatar-md" style={{ background: avatarColor(c.firstName) }}>
                {c.avatarUrl
                  ? <img src={c.avatarUrl} alt="" />
                  : (c.firstName?.[0] || '?').toUpperCase()
                }
              </div>

              <div className={`match-badge ${matchClass(c.overallScore)}`}>
                <span className="match-pct">{Math.round(c.overallScore * 100)}%</span>
                <span className="match-label">match</span>
              </div>

              <div className="cand-info">
                <div className="cand-name">{c.firstName} {c.lastName}</div>
                {c.bio && <div className="cand-bio">{c.bio}</div>}

                <div className="chips-row">
                  {c.commonSlots?.length > 0 && (
                    <span className="chip chip-time">
                      {c.commonSlots.length} common slot{c.commonSlots.length > 1 ? 's' : ''}
                    </span>
                  )}
                  {c.commonCourses?.map((course, i) => (
                    <span key={i} className="chip chip-course">{course}</span>
                  ))}
                </div>

                <div className="cand-actions">
                  <button
                    type="button"
                    className="btn btn-primary btn-sm"
                    disabled={sending === c.userId}
                    onClick={() => handleSend(c.userId)}
                  >
                    {sending === c.userId ? 'Sending...' : 'Connect'}
                  </button>
                  <button
                    type="button"
                    className="btn btn-secondary btn-sm"
                    onClick={() => setOpenMsg(openMsg === c.userId ? null : c.userId)}
                  >
                    {openMsg === c.userId ? 'Hide message' : 'Add message'}
                  </button>
                </div>

                {openMsg === c.userId && (
                  <input
                    className="cand-msg-input"
                    placeholder="Write a short message (optional)..."
                    value={messages[c.userId] || ''}
                    onChange={(e) => setMessages((m) => ({ ...m, [c.userId]: e.target.value }))}
                    maxLength={500}
                  />
                )}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default Candidates
