import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { apiListCourses } from '../../api'
import { useAuth } from '../../context/useAuth'

const levelEmoji = (level) => {
  if (!level) return '📘'
  const l = level.toLowerCase()
  if (l.includes('beginner') || l.includes('intro')) return '🌱'
  if (l.includes('advanced') || l.includes('expert')) return '🔥'
  return '📘'
}

const CourseList = () => {
  const { profile } = useAuth()
  const [courses, setCourses] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [subject, setSubject] = useState('')
  const [level, setLevel] = useState('')

  const load = async () => {
    setLoading(true)
    setError('')
    try {
      const data = await apiListCourses({
        subject: subject || undefined,
        level: level || undefined,
        limit: 50,
      })
      const all = Array.isArray(data) ? data : []
      setCourses(profile ? all.filter((c) => c.ownerUserId === profile.id) : all)
    } catch (e) {
      setError(e.error || 'Failed to load courses')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [subject, level, profile?.id])

  return (
    <div className="courses-page">
      <div className="card">
        <div className="courses-head">
          <div>
            <div className="courses-title">My Courses</div>
            <div className="courses-sub">Courses you're studying</div>
          </div>
          <Link to="/courses/new" className="btn btn-primary btn-sm">+ New course</Link>
        </div>

        <div style={{ padding: '0 20px 12px', display: 'flex', gap: '8px' }}>
          <input
            className="profile-input"
            placeholder="Filter by subject"
            value={subject}
            onChange={(e) => setSubject(e.target.value)}
            style={{ flex: 1 }}
          />
          <input
            className="profile-input"
            placeholder="Filter by level"
            value={level}
            onChange={(e) => setLevel(e.target.value)}
            style={{ flex: 1 }}
          />
        </div>

        {error && <div style={{ padding: '0 20px 12px', color: 'var(--danger)', fontSize: '13px' }}>{error}</div>}
        {loading && <div className="loading-state">Loading courses...</div>}

        {!loading && courses.length === 0 && (
          <div className="empty-state">
            No courses yet.{' '}
            <Link to="/courses/new" style={{ color: 'var(--primary)', fontWeight: 600 }}>Create one</Link>
          </div>
        )}

        {!loading && courses.map((c) => (
          <Link key={c.id} to={`/courses/${c.id}`} className="course-item" style={{ display: 'flex', textDecoration: 'none' }}>
            <div className="course-item-icon">{levelEmoji(c.level)}</div>
            <div className="course-item-info">
              <div className="course-item-title">{c.title}</div>
              <div className="course-item-meta">
                {[c.subject, c.level].filter(Boolean).join(' · ')}
              </div>
              {c.description && (
                <div className="course-item-footer"
                  style={{ overflow: 'hidden', display: '-webkit-box', WebkitLineClamp: 1, WebkitBoxOrient: 'vertical' }}>
                  {c.description}
                </div>
              )}
            </div>
            <span style={{ fontSize: '18px', color: 'var(--muted)', alignSelf: 'center', marginLeft: '8px' }}>›</span>
          </Link>
        ))}
      </div>
    </div>
  )
}

export default CourseList
