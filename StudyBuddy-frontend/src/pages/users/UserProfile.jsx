import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useToast } from '../../context/ToastContext'
import { useAuth } from '../../context/useAuth'
import { apiGetUserById, apiGetUserRating, apiListReviewsForUser } from '../../api'
import { avatarColor } from '../../utils/avatar'

export default function UserProfile() {
  const { id }       = useParams()
  const navigate     = useNavigate()
  const toast        = useToast()
  const { profile }  = useAuth()

  const [user,    setUser]    = useState(null)
  const [rating,  setRating]  = useState(null)
  const [reviews, setReviews] = useState([])
  const [loading, setLoading] = useState(true)
  const [copied,  setCopied]  = useState(false)

  const isMe = profile?.id === id

  useEffect(() => {
    if (isMe) { navigate('/profile', { replace: true }); return }

    const load = async () => {
      setLoading(true)
      try {
        const [u, r, rv] = await Promise.allSettled([
          apiGetUserById(id),
          apiGetUserRating(id),
          apiListReviewsForUser(id),
        ])
        if (u.status === 'fulfilled')  setUser(u.value)
        if (r.status === 'fulfilled')  setRating(r.value)
        if (rv.status === 'fulfilled') setReviews(rv.value?.items ?? rv.value ?? [])
      } catch {
        toast.error('Failed to load profile')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [id])

  const copyId = async () => {
    try {
      await navigator.clipboard.writeText(id)
      setCopied(true)
      toast.success('User ID copied!')
      setTimeout(() => setCopied(false), 2000)
    } catch {
      toast.error('Copy failed')
    }
  }

  if (loading) return <div className="loading-state" style={{ marginTop: 40 }}>Loading profile…</div>

  if (!user) return (
    <div className="empty-state" style={{ marginTop: 40 }}>
      User not found.{' '}
      <button className="btn btn-secondary btn-sm" onClick={() => navigate(-1)}>Go back</button>
    </div>
  )

  const name    = `${user.firstName ?? ''} ${user.lastName ?? ''}`.trim()
  const initial = (user.firstName?.[0] || '?').toUpperCase()
  const avg     = rating?.averageRating

  return (
    <div style={{ maxWidth: 640, margin: '0 auto' }}>

      {/* Back */}
      <button
        className="btn btn-ghost btn-sm"
        style={{ marginBottom: 12 }}
        onClick={() => navigate(-1)}
      >
        ← Back
      </button>

      {/* Header card */}
      <div className="card" style={{ marginBottom: 14, overflow: 'hidden' }}>
        {/* Cover */}
        <div style={{
          height: clamp(80, 100),
          background: 'linear-gradient(135deg, #1e3a8a 0%, #3b82f6 100%)',
        }} />

        <div style={{ padding: '0 20px 20px' }}>
          {/* Avatar row */}
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', marginTop: -32, marginBottom: 12, flexWrap: 'wrap', gap: 8 }}>
            <div
              className="avatar"
              style={{
                width: 72, height: 72,
                fontSize: 26, fontWeight: 700,
                border: '3px solid #fff',
                background: user.avatarUrl ? undefined : avatarColor(user.firstName),
              }}
            >
              {user.avatarUrl ? <img src={user.avatarUrl} alt={name} /> : initial}
            </div>

            {/* Rating */}
            {avg != null && (
              <div style={{
                display: 'flex', alignItems: 'center', gap: 5,
                background: '#fefce8', border: '1px solid #fde68a',
                borderRadius: 8, padding: '6px 12px',
              }}>
                <span style={{ fontSize: 18, lineHeight: 1 }}>⭐</span>
                <span style={{ fontWeight: 800, fontSize: 16, color: '#92400e' }}>
                  {avg.toFixed(1)}
                </span>
                {rating.reviewCount != null && (
                  <span style={{ fontSize: 12, color: 'var(--muted)' }}>
                    ({rating.reviewCount})
                  </span>
                )}
              </div>
            )}
          </div>

          {/* Name */}
          <div style={{ fontSize: 22, fontWeight: 800, color: 'var(--text)' }}>{name}</div>
          {user.telegramTag && (
            <div style={{ fontSize: 13, color: 'var(--muted)', marginTop: 3 }}>
              @{user.telegramTag}
            </div>
          )}

          {/* Bio */}
          {user.bio && (
            <div style={{ fontSize: 14, color: 'var(--muted)', marginTop: 10, lineHeight: 1.6 }}>
              {user.bio}
            </div>
          )}

          {/* Copy ID button */}
          <div style={{
            marginTop: 16,
            padding: '12px 14px',
            background: 'var(--bg)',
            borderRadius: 10,
            border: '1px solid var(--border)',
            display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 12,
          }}>
            <div>
              <div style={{ fontSize: 11, fontWeight: 600, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: '0.5px', marginBottom: 3 }}>
                User ID
              </div>
              <div style={{ fontFamily: 'monospace', fontSize: 12, color: 'var(--text)', wordBreak: 'break-all' }}>
                {id}
              </div>
            </div>
            <button
              className={`btn btn-sm ${copied ? 'btn-primary' : 'btn-secondary'}`}
              style={{ flexShrink: 0 }}
              onClick={copyId}
            >
              {copied ? '✓ Copied' : 'Copy ID'}
            </button>
          </div>
        </div>
      </div>

      {/* Interests */}
      {user.interests?.length > 0 && (
        <div className="card" style={{ marginBottom: 14, padding: '16px 20px' }}>
          <div style={{ fontSize: 15, fontWeight: 700, marginBottom: 12 }}>Interests</div>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
            {user.interests.map((i) => (
              <span key={i.id} className="chip chip-int">{i.name}</span>
            ))}
          </div>
        </div>
      )}

      {/* Reviews */}
      {reviews.length > 0 && (
        <div className="card" style={{ padding: '16px 20px' }}>
          <div style={{ fontSize: 15, fontWeight: 700, marginBottom: 12 }}>
            Reviews ({reviews.length})
          </div>
          {reviews.map((rv) => (
            <div key={rv.id} style={{
              padding: '12px 0',
              borderBottom: '1px solid var(--border)',
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 6 }}>
                {'⭐'.repeat(rv.rating ?? 0)}
                <span style={{ fontSize: 12, color: 'var(--muted)' }}>
                  {rv.createdAt ? new Date(rv.createdAt).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : ''}
                </span>
              </div>
              {rv.comment && (
                <div style={{ fontSize: 14, color: 'var(--text)', lineHeight: 1.5 }}>
                  {rv.comment}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

function clamp(min, max) {
  return `clamp(${min}px, 12vw, ${max}px)`
}
