import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../../context/useAuth'
import { apiListCourses, apiGetMatchRequests, apiGetCandidates, apiRespondMatchRequest, apiGetMyInterests } from '../../api'
import { useToast } from '../../context/ToastContext'

const matchClass = (score) => {
  if (score >= 0.7) return 'match-green'
  if (score >= 0.4) return 'match-amber'
  return 'match-gray'
}

const avatarBg = 'linear-gradient(135deg,#60a5fa 0%,#3b82f6 100%)'

const Dashboard = () => {
  const { profile } = useAuth()
  const toast = useToast()
  const [stats, setStats] = useState({ courses: null, matches: null, pending: null })
  const [candidates, setCandidates] = useState([])
  const [pendingReqs, setPendingReqs] = useState([])
  const [interests, setInterests] = useState([])
  const [acting, setActing] = useState(null)

  useEffect(() => {
    const load = async () => {
      try {
        const [coursesData, acceptedData, pendingData, candsData, interestsData] = await Promise.all([
          apiListCourses({ limit: 100 }),
          apiGetMatchRequests({ status: 'accepted', limit: 100 }),
          apiGetMatchRequests({ status: 'pending', limit: 100 }),
          apiGetCandidates(3),
          apiGetMyInterests(),
        ])
        const allCourses = Array.isArray(coursesData) ? coursesData : []
        const myCourses = profile ? allCourses.filter((c) => c.ownerUserId === profile.id) : allCourses
        const accepted = acceptedData.items ?? []
        const pending = pendingData.items ?? []
        const myPending = pending.filter((r) => r.receiverId === profile?.id)

        setStats({ courses: myCourses.length, matches: accepted.length, pending: myPending.length })
        setCandidates(candsData.items ?? [])
        setPendingReqs(myPending.slice(0, 3))
        setInterests((interestsData.items ?? []).slice(0, 8))
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
      toast.success(accept ? 'Request accepted!' : 'Declined')
      setPendingReqs((prev) => prev.filter((r) => r.id !== id))
      setStats((s) => ({ ...s, pending: Math.max(0, (s.pending ?? 1) - 1) }))
    } catch (e) {
      toast.error(e.error || 'Failed')
    } finally {
      setActing(null)
    }
  }

  const firstName = profile?.firstName || 'student'
  const initial = (profile?.firstName?.[0] || '?').toUpperCase()

  return (
    <div className="dash-layout">
      {/* ── Left: user card ── */}
      <aside>
        <div className="card user-card">
          <div className="user-card-banner" />
          <div className="user-card-body">
            <div className="user-card-ava-wrap">
              <div className="user-card-ava" style={{ background: avatarBg }}>
                {profile?.avatarUrl
                  ? <img src={profile.avatarUrl} alt="" />
                  : initial
                }
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
              <div className="avatar avatar-md" style={{ background: avatarBg }}>
                {c.avatarUrl ? <img src={c.avatarUrl} alt="" /> : (c.firstName?.[0] || '?').toUpperCase()}
              </div>
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
        {/* Pending requests widget */}
        <div className="card widget">
          <div className="widget-header">
            <span className="widget-title">Incoming requests</span>
            <Link to="/matching/requests" className="widget-link">View all</Link>
          </div>

          {pendingReqs.length === 0
            ? <div className="widget-empty">No pending requests.</div>
            : pendingReqs.map((r) => (
                <div key={r.id} className="req-mini">
                  <div className="avatar avatar-sm" style={{ background: avatarBg }}>?</div>
                  <div className="req-mini-info">
                    <div className="req-mini-name">New request</div>
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
              ))
          }
        </div>

        {/* Interests widget */}
        {interests.length > 0 && (
          <div className="card widget">
            <div className="widget-header">
              <span className="widget-title">My interests</span>
              <Link to="/interests" className="widget-link">Edit</Link>
            </div>
            <div className="int-tags">
              {interests.map((i) => (
                <span key={i.ID} className="chip chip-int">{i.Name}</span>
              ))}
            </div>
          </div>
        )}

        {/* Quick links widget */}
        <div className="card widget">
          <div className="widget-header">
            <span className="widget-title">Quick links</span>
          </div>
          <div style={{ padding: '0 14px 12px', display: 'flex', flexDirection: 'column', gap: '4px' }}>
            {[
              ['/availability', 'Set availability'],
              ['/courses/new', 'Add a course'],
              ['/matching/partners', 'My partners'],
            ].map(([to, label]) => (
              <Link key={to} to={to} className="sidebar-nav-link" style={{ fontSize: '12px' }}>{label}</Link>
            ))}
          </div>
        </div>
      </aside>
    </div>
  )
}

export default Dashboard
