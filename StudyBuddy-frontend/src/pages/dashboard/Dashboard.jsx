import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../../context/useAuth'
import { useToast } from '../../context/ToastContext'
import {
  apiListCourses, apiGetMatchRequests, apiGetCandidates,
  apiRespondMatchRequest, apiGetMyInterests, apiGetUserById, apiGetSlots,
} from '../../api'
import { avatarColor } from '../../utils/avatar'

// backend returns 0-indexed: 0=Mon, 1=Tue, 2=Wed, 3=Thu, 4=Fri, 5=Sat, 6=Sun
const DAY_SHORT  = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
const DAY_COLORS = ['#3b82f6','#8b5cf6','#10b981','#f59e0b','#6366f1','#ec4899','#ef4444']

const matchClass = (score) => {
  if (score >= 0.7) return 'match-green'
  if (score >= 0.4) return 'match-amber'
  return 'match-gray'
}

const Avatar = ({ name, url, size = 'avatar-md' }) => {
  const letter = (name || '?')[0].toUpperCase()
  const bg = avatarColor(name)
  return (
    <div className={`avatar ${size}`} style={{ background: bg }}>
      {url ? <img src={url} alt="" /> : letter}
    </div>
  )
}

const Dashboard = () => {
  const { profile } = useAuth()
  const toast = useToast()
  const [stats, setStats] = useState({ courses: null, matches: null, pending: null })
  const [candidates, setCandidates] = useState([])
  const [pendingReqs, setPendingReqs] = useState([])
  const [reqUsers, setReqUsers] = useState({})
  const [interests, setInterests] = useState([])
  const [slots, setSlots] = useState([])
  const [acting, setActing] = useState(null)

  useEffect(() => {
    const load = async () => {
      try {
        const [coursesData, acceptedData, pendingData, candsData, interestsData, slotsData] = await Promise.all([
          apiListCourses({ limit: 100 }),
          apiGetMatchRequests({ status: 'accepted', limit: 100 }),
          apiGetMatchRequests({ status: 'pending', limit: 100 }),
          apiGetCandidates(3),
          apiGetMyInterests(),
          apiGetSlots(),
        ])

        const allCourses = Array.isArray(coursesData) ? coursesData : []
        const myCourses = profile ? allCourses.filter((c) => c.ownerUserId === profile.id) : allCourses
        const accepted = acceptedData.items ?? []
        const pending = pendingData.items ?? []
        const myPending = pending.filter((r) => r.receiverId === profile?.id)

        setStats({ courses: myCourses.length, matches: accepted.length, pending: myPending.length })
        setCandidates(candsData.items ?? [])
        setInterests((interestsData.items ?? []).slice(0, 10))
        setSlots((slotsData.items ?? []).sort((a, b) => a.dayOfWeek - b.dayOfWeek))

        const top = myPending.slice(0, 4)
        setPendingReqs(top)

        // fetch names for request senders
        const userResults = await Promise.allSettled(
          top.map((r) => apiGetUserById(r.requesterId).then((u) => [r.requesterId, u]))
        )
        const cache = {}
        for (const res of userResults) {
          if (res.status === 'fulfilled') {
            const [id, user] = res.value
            cache[id] = user
          }
        }
        setReqUsers(cache)
      } catch {
        // non-critical
      }
    }
    load()
  }, [profile?.id])

  const handleRespond = async (id, accept) => {
    setActing(id)
    try {
      await apiRespondMatchRequest(id, accept)
      toast.success(accept ? 'Accepted!' : 'Declined')
      setPendingReqs((prev) => prev.filter((r) => r.id !== id))
      setStats((s) => ({ ...s, pending: Math.max(0, (s.pending ?? 1) - 1) }))
    } catch (e) {
      toast.error(e.error || 'Failed')
    } finally {
      setActing(null)
    }
  }

  const initial = (profile?.firstName?.[0] || '?').toUpperCase()

  return (
    <div className="dash-layout">
      {/* ── Left: user card ── */}
      <aside>
        <div className="card user-card">
          <div className="user-card-banner" />
          <div className="user-card-body">
            <div className="user-card-ava-wrap">
              <div
                className="user-card-ava"
                style={{ background: avatarColor(profile?.firstName || profile?.email) }}
              >
                {profile?.avatarUrl ? <img src={profile.avatarUrl} alt="" /> : initial}
              </div>
            </div>
            <div className="user-card-name">{profile?.firstName} {profile?.lastName}</div>
            <div className="user-card-role">{profile?.email}</div>
            {profile?.bio && (
              <div style={{ fontSize: '11px', color: 'var(--muted)', marginTop: '6px', lineHeight: 1.4 }}>
                {profile.bio}
              </div>
            )}
            <div className="user-stats">
              <div className="user-stat-row">
                <span>Courses</span>
                <span className="user-stat-val">{stats.courses ?? '—'}</span>
              </div>
              <div className="user-stat-row">
                <span>Partners</span>
                <span className="user-stat-val">{stats.matches ?? '—'}</span>
              </div>
              <div className="user-stat-row">
                <span>Pending</span>
                <span className="user-stat-val">{stats.pending ?? '—'}</span>
              </div>
            </div>
            <nav className="sidebar-nav">
              {[
                ['/profile', 'Edit profile'],
                ['/interests', 'My interests'],
                ['/availability', 'Availability'],
                ['/courses', 'My courses'],
              ].map(([to, label]) => (
                <Link key={to} to={to} className="sidebar-nav-link">{label}</Link>
              ))}
            </nav>
          </div>
        </div>
      </aside>

      {/* ── Center: recommended partners ── */}
      <main>
        <div className="card" style={{ marginBottom: '14px' }}>
          <div className="section-header">
            <div>
              <div className="section-title">Recommended partners</div>
              <div className="section-sub">Based on your interests, courses and schedule</div>
            </div>
            <Link to="/matching/candidates" className="section-link">See all</Link>
          </div>

          {candidates.length === 0 && (
            <div className="empty-state">
              No candidates yet. Fill in your interests and availability to get matches.
            </div>
          )}

          {candidates.map((c) => (
            <div key={c.userId} className="rec-card">
              <Avatar name={c.firstName} url={c.avatarUrl} size="avatar-md" />
              <div className={`match-badge ${matchClass(c.overallScore)}`}>
                <span className="match-pct">{Math.round(c.overallScore * 100)}%</span>
                <span className="match-label">match</span>
              </div>
              <div className="rec-info">
                <div className="rec-name">{c.firstName} {c.lastName}</div>
                <div className="rec-role">
                  {c.commonCourses?.length > 0 && `${c.commonCourses.length} shared course${c.commonCourses.length > 1 ? 's' : ''}`}
                  {c.commonCourses?.length > 0 && c.commonSlots?.length > 0 && ' · '}
                  {c.commonSlots?.length > 0 && `${c.commonSlots.length} common slot${c.commonSlots.length > 1 ? 's' : ''}`}
                </div>
                {c.bio && <div className="rec-bio">{c.bio}</div>}
                <div className="rec-actions">
                  <Link to="/matching/candidates" className="btn btn-primary btn-sm">Connect</Link>
                </div>
              </div>
            </div>
          ))}
        </div>
      </main>

      {/* ── Right: widgets ── */}
      <aside>
        {/* Incoming requests widget */}
        <div className="card widget">
          <div className="widget-header">
            <span className="widget-title">Incoming requests</span>
            <Link to="/matching/requests" className="widget-link">All</Link>
          </div>

          {pendingReqs.length === 0
            ? <div className="widget-empty">No pending requests.</div>
            : pendingReqs.map((r) => {
                const u = reqUsers[r.requesterId]
                const name = u ? `${u.firstName} ${u.lastName}` : '...'
                return (
                  <div key={r.id} className="req-mini">
                    <Avatar name={u?.firstName || r.requesterId} url={u?.avatarUrl} size="avatar-sm" />
                    <div className="req-mini-info">
                      <div className="req-mini-name">{name}</div>
                      <div className="req-mini-msg">{r.message || 'No message'}</div>
                    </div>
                    <div className="req-mini-btns">
                      <button
                        type="button"
                        className="req-mini-btn req-mini-accept"
                        disabled={acting === r.id}
                        onClick={() => handleRespond(r.id, true)}
                        title="Accept"
                      >✓</button>
                      <button
                        type="button"
                        className="req-mini-btn req-mini-decline"
                        disabled={acting === r.id}
                        onClick={() => handleRespond(r.id, false)}
                        title="Decline"
                      >✕</button>
                    </div>
                  </div>
                )
              })
          }
        </div>

        {/* Schedule widget */}
        <div className="card widget">
          <div className="widget-header">
            <span className="widget-title">My schedule</span>
            <Link to="/availability" className="widget-link">Edit</Link>
          </div>

          {slots.length === 0
            ? <div className="widget-empty">No slots added yet.</div>
            : slots.map((s) => (
                <div key={s.id} className="sched-item">
                  <div
                    className="sched-dot"
                    style={{ background: DAY_COLORS[s.dayOfWeek] || '#94a3b8' }}
                  >
                    {DAY_SHORT[s.dayOfWeek]}
                  </div>
                  <div>
                    <div className="sched-time">{s.startTime} – {s.endTime}</div>
                  </div>
                </div>
              ))
          }
        </div>

        {/* Interests widget */}
        {interests.length > 0 && (
          <div className="card widget">
            <div className="widget-header">
              <span className="widget-title">My interests</span>
              <Link to="/interests" className="widget-link">+ Add</Link>
            </div>
            <div className="int-tags">
              {interests.map((i) => (
                <span key={i.ID} className="chip chip-int">{i.Name}</span>
              ))}
            </div>
          </div>
        )}
      </aside>
    </div>
  )
}

export default Dashboard
