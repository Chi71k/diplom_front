import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useToast } from '../../context/ToastContext'
import { useAuth } from '../../context/useAuth'
import { apiGetUserById, apiGetUserRating, apiListReviewsForUser, apiListCourses, apiGetUserPoints } from '../../api'
import { avatarColor } from '../../utils/avatar'

export default function UserProfile() {
  const { id }       = useParams()
  const navigate     = useNavigate()
  const toast        = useToast()
  const { profile }  = useAuth()

  const [user,    setUser]    = useState(null)
  const [rating,  setRating]  = useState(null)
  const [points,  setPoints]  = useState(null)
  const [reviews, setReviews] = useState([])
  const [courses, setCourses] = useState([])
  const [loading, setLoading] = useState(true)

  const isMe = profile?.id === id

  useEffect(() => {
    if (isMe) { navigate('/profile', { replace: true }); return }

    const load = async () => {
      setLoading(true)
      try {
        const [u, r, pts, rv, cs] = await Promise.allSettled([
          apiGetUserById(id),
          apiGetUserRating(id),
          apiGetUserPoints(id),
          apiListReviewsForUser(id),
          apiListCourses({ limit: 200 }),
        ])
        if (u.status === 'fulfilled')   setUser(u.value)
        if (r.status === 'fulfilled')   setRating(r.value)
        if (pts.status === 'fulfilled') setPoints(pts.value)
        if (rv.status === 'fulfilled')  setReviews(rv.value?.items ?? rv.value ?? [])
        if (cs.status === 'fulfilled') {
          const all = Array.isArray(cs.value) ? cs.value : (cs.value?.items ?? [])
          setCourses(all.filter(c => c.ownerUserId === id))
        }
      } catch {
        toast.error('Failed to load profile')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [id])

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

      <button
        className="btn btn-ghost btn-sm"
        style={{ marginBottom: 12 }}
        onClick={() => navigate(-1)}
      >
        ← Back
      </button>

      <div className="card" style={{ marginBottom: 14, overflow: 'hidden' }}>
        <div style={{
          height: clamp(80, 100),
          background: 'linear-gradient(135deg, #1e3a8a 0%, #3b82f6 100%)',
        }} />

        <div style={{ padding: '0 20px 20px' }}>
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

            <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
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
              {points?.totalPoints != null && (
                <div style={{
                  display: 'flex', alignItems: 'center', gap: 5,
                  background: 'var(--blue-50)', border: '1px solid #bfdbfe',
                  borderRadius: 8, padding: '6px 12px',
                }}>
                  <span style={{ fontSize: 16, lineHeight: 1 }}>★</span>
                  <span style={{ fontWeight: 800, fontSize: 15, color: 'var(--primary)' }}>
                    {points.totalPoints}
                  </span>
                  <span style={{ fontSize: 12, color: 'var(--muted)' }}>pts</span>
                </div>
              )}
            </div>
          </div>

          <div style={{ fontSize: 22, fontWeight: 800, color: 'var(--text)' }}>{name}</div>
          {user.telegramTag && (
            <div style={{ fontSize: 13, marginTop: 4 }}>
              <a
                href={`https://t.me/${user.telegramTag.replace(/^@/, '')}`}
                target="_blank"
                rel="noopener noreferrer"
                style={{ color: '#0088cc', textDecoration: 'none', fontWeight: 500 }}
              >
                @{user.telegramTag.replace(/^@/, '')}
              </a>
            </div>
          )}

          {/* Bio */}
          {user.bio && (
            <div style={{ fontSize: 14, color: 'var(--muted)', marginTop: 10, lineHeight: 1.6 }}>
              {user.bio}
            </div>
          )}
        </div>
      </div>

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

      {courses.length > 0 && (
        <div className="card" style={{ marginBottom: 14, padding: '16px 20px' }}>
          <div style={{ fontSize: 15, fontWeight: 700, marginBottom: 12 }}>Courses ({courses.length})</div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {courses.map((c) => (
              <div key={c.id} style={{
                padding: '10px 14px',
                background: 'var(--bg)',
                borderRadius: 8,
                border: '1px solid var(--border)',
              }}>
                <div style={{ fontWeight: 600, fontSize: 14, color: 'var(--text)' }}>{c.title}</div>
                {(c.subject || c.level) && (
                  <div style={{ display: 'flex', gap: 6, marginTop: 5 }}>
                    {c.subject && <span className="chip chip-int" style={{ fontSize: 11 }}>{c.subject}</span>}
                    {c.level   && <span className="chip chip-int" style={{ fontSize: 11 }}>{c.level}</span>}
                  </div>
                )}
                {c.description && (
                  <div style={{ fontSize: 13, color: 'var(--muted)', marginTop: 5 }}>
                    {c.description.length > 100 ? c.description.slice(0, 100) + '…' : c.description}
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

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
