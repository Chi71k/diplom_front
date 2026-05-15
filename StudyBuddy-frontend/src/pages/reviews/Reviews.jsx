import { useState, useEffect } from 'react'
import { useToast } from '../../context/ToastContext'
import { useAuth } from '../../context/useAuth'
import {
  apiListReviewsForUser,
  apiGetMatchRequests,
  apiGetUserById,
  apiCreateReview,
} from '../../api'
import { avatarColor } from '../../utils/avatar'

const Stars = ({ value, onChange }) => (
  <div style={{ display: 'flex', gap: '2px' }}>
    {[1, 2, 3, 4, 5].map((n) => (
      <button
        key={n}
        type="button"
        onClick={() => onChange?.(n)}
        style={{
          background: 'none',
          border: 'none',
          cursor: onChange ? 'pointer' : 'default',
          fontSize: '24px',
          color: n <= value ? '#f59e0b' : '#d1d5db',
          padding: '0 1px',
          lineHeight: 1,
        }}
      >
        ★
      </button>
    ))}
  </div>
)

const Reviews = () => {
  const toast = useToast()
  const { profile } = useAuth()
  const [myReviews, setMyReviews] = useState([])
  const [partners, setPartners] = useState([])
  const [loading, setLoading] = useState(true)
  const [tab, setTab] = useState('received')
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ revieweeId: '', matchId: '', rating: 5, comment: '' })
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (!profile?.id) return
    const load = async () => {
      setLoading(true)
      try {
        const [reviewsData, requestsData] = await Promise.all([
          apiListReviewsForUser(profile.id),
          apiGetMatchRequests({ status: 'accepted', limit: 100 }),
        ])
        setMyReviews(reviewsData ?? [])

        const items = requestsData.items ?? []
        const pairs = items.map((r) => ({
          matchId: r.id,
          partnerId: r.requesterId === profile.id ? r.receiverId : r.requesterId,
        }))

        const results = await Promise.allSettled(
          pairs.map(({ partnerId }) => apiGetUserById(partnerId))
        )

        // Keep all partners even if name lookup fails — use ID as fallback
        const enriched = pairs.map(({ matchId, partnerId }, i) => {
          const user = results[i].status === 'fulfilled'
            ? results[i].value
            : { id: partnerId, firstName: partnerId?.slice(0, 8) ?? '?', lastName: '' }
          return { ...user, matchId }
        })
        setPartners(enriched)
      } catch (e) {
        toast.error(e.error || 'Failed to load reviews')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [profile?.id])

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (!form.revieweeId) { toast.error('Select a partner first'); return }
    setSubmitting(true)
    try {
      const rev = await apiCreateReview({
        revieweeId: form.revieweeId,
        matchId: form.matchId,
        rating: form.rating,
        comment: form.comment,
      })
      toast.success('Review submitted!')
      setMyReviews((prev) => [rev, ...prev])
      setShowForm(false)
      setForm({ revieweeId: '', matchId: '', rating: 5, comment: '' })
    } catch (e) {
      toast.error(e.error || 'Failed to submit review')
    } finally {
      setSubmitting(false)
    }
  }

  const received = myReviews.filter((r) => r.revieweeId === profile?.id)
  const given    = myReviews.filter((r) => r.reviewerId === profile?.id)
  const shown    = tab === 'received' ? received : given

  const fmt = (iso) =>
    new Date(iso).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })

  const partnerName = (id) => {
    const p = partners.find((x) => x.id === id)
    if (!p) return id?.slice(0, 8) ?? id
    return `${p.firstName} ${p.lastName}`.trim() || id?.slice(0, 8)
  }

  return (
    <div className="requests-page">
      <div className="card">

        {/* Header */}
        <div className="requests-head" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <div className="requests-title">Reviews</div>
            <div className="requests-sub">Rate your study partners and see what others say about you</div>
          </div>
          <button
            className={`btn btn-sm ${showForm ? 'btn-secondary' : 'btn-primary'}`}
            onClick={() => setShowForm((v) => !v)}
          >
            {showForm ? 'Cancel' : '+ Write review'}
          </button>
        </div>

        {/* Write review form */}
        {showForm && (
          <form
            onSubmit={handleSubmit}
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
                Partner
              </label>
              <select
                className="form-input"
                style={{
                  width: '100%',
                  padding: '9px 12px',
                  border: '1px solid var(--border)',
                  borderRadius: '8px',
                  background: '#fff',
                  fontSize: '14px',
                  outline: 'none',
                }}
                required
                value={form.revieweeId}
                onChange={(e) => {
                  const p = partners.find((x) => x.id === e.target.value)
                  setForm((f) => ({ ...f, revieweeId: e.target.value, matchId: p?.matchId || '' }))
                }}
              >
                <option value="">Select partner to review…</option>
                {partners.map((p) => (
                  <option key={p.id} value={p.id}>
                    {`${p.firstName} ${p.lastName}`.trim() || p.id?.slice(0, 8)}
                  </option>
                ))}
              </select>
              {partners.length === 0 && !loading && (
                <div style={{ fontSize: '12px', color: 'var(--muted)', marginTop: '6px' }}>
                  No accepted partners yet — connect with someone first.
                </div>
              )}
            </div>

            <div>
              <label style={{ fontSize: '12px', color: 'var(--muted)', marginBottom: '6px', display: 'block' }}>
                Rating
              </label>
              <Stars value={form.rating} onChange={(r) => setForm((f) => ({ ...f, rating: r }))} />
            </div>

            <div>
              <label style={{ fontSize: '12px', color: 'var(--muted)', marginBottom: '6px', display: 'block' }}>
                Comment (optional)
              </label>
              <textarea
                style={{
                  width: '100%',
                  padding: '9px 12px',
                  border: '1px solid var(--border)',
                  borderRadius: '8px',
                  fontSize: '14px',
                  outline: 'none',
                  resize: 'vertical',
                  minHeight: '80px',
                  background: '#fff',
                }}
                placeholder="Share your experience with this partner…"
                rows={3}
                value={form.comment}
                onChange={(e) => setForm((f) => ({ ...f, comment: e.target.value }))}
                maxLength={1000}
              />
            </div>

            <button
              type="submit"
              className="btn btn-primary"
              style={{ alignSelf: 'flex-start' }}
              disabled={submitting}
            >
              {submitting ? 'Submitting…' : 'Submit review'}
            </button>
          </form>
        )}

        {/* Tabs */}
        <div className="tabs">
          {[
            { key: 'received', label: `Received (${received.length})` },
            { key: 'given',    label: `Given (${given.length})` },
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

        {loading && <div className="loading-state">Loading reviews…</div>}

        {!loading && shown.length === 0 && (
          <div className="empty-state">No {tab} reviews yet.</div>
        )}

        {!loading && shown.map((r) => {
          const otherId = tab === 'received' ? r.reviewerId : r.revieweeId
          const name = partnerName(otherId)
          return (
            <div key={r.id} className="req-card">
              <div className="req-card-main">
                <div
                  className="avatar avatar-sm"
                  style={{ background: avatarColor(otherId), flexShrink: 0 }}
                >
                  {name?.[0]?.toUpperCase() ?? '?'}
                </div>
                <div className="req-card-info">
                  <div style={{ display: 'flex', alignItems: 'center', gap: '10px', flexWrap: 'wrap' }}>
                    <span className="req-card-name">{name}</span>
                    <Stars value={r.rating} />
                    <span style={{ fontSize: '12px', color: 'var(--muted)', marginLeft: 'auto' }}>
                      {fmt(r.createdAt)}
                    </span>
                  </div>
                  <div className="req-card-role">
                    {tab === 'received' ? 'reviewed you' : `you reviewed`}
                  </div>
                  {r.comment && (
                    <div className="req-card-msg" style={{ marginTop: '8px' }}>
                      "{r.comment}"
                    </div>
                  )}
                </div>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

export default Reviews
