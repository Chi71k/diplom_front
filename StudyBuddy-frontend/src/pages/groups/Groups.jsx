import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useToast } from '../../context/ToastContext'
import { apiListMyGroups } from '../../api'

const groupColors = ['#dbeafe', '#dcfce7', '#fef9c3', '#fce7f3', '#ede9fe', '#ffedd5']
const groupTextColors = ['#1d4ed8', '#15803d', '#a16207', '#be185d', '#7c3aed', '#c2410c']

const groupColorPair = (name) => {
  const idx = (name?.charCodeAt(0) ?? 0) % groupColors.length
  return { bg: groupColors[idx], color: groupTextColors[idx] }
}

const Groups = () => {
  const toast = useToast()
  const [groups, setGroups]   = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const load = async () => {
      setLoading(true)
      try {
        const data = await apiListMyGroups()
        setGroups(data.items ?? [])
      } catch (e) {
        toast.error(e.error || 'Failed to load groups')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [])

  return (
    <div className="requests-page">
      <div className="card">

        {/* Header */}
        <div className="requests-head" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <div className="requests-title">Study Groups</div>
            <div className="requests-sub">Collaborate with multiple partners in a group</div>
          </div>
          <Link to="/groups/new" className="btn btn-primary btn-sm">+ New group</Link>
        </div>

        {loading && <div className="loading-state">Loading groups…</div>}

        {!loading && groups.length === 0 && (
          <div className="empty-state">
            No groups yet.{' '}
            <Link to="/groups/new" style={{ color: 'var(--primary)', fontWeight: 600 }}>
              Create one!
            </Link>
          </div>
        )}

        {!loading && groups.map((g) => {
          const { bg, color } = groupColorPair(g.name)
          const memberCount   = g.members?.length ?? 0
          const courseCount   = g.courseIds?.length ?? 0
          return (
            <Link key={g.id} to={`/groups/${g.id}`} style={{ textDecoration: 'none' }}>
              <div
                className="req-card"
                style={{ cursor: 'pointer', transition: 'background .12s' }}
                onMouseEnter={(e) => e.currentTarget.style.background = '#fafafa'}
                onMouseLeave={(e) => e.currentTarget.style.background = ''}
              >
                <div className="req-card-main" style={{ alignItems: 'center' }}>
                  <div style={{
                    width: '44px', height: '44px', borderRadius: '12px',
                    background: bg, color, flexShrink: 0,
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                    fontSize: '20px', fontWeight: 800,
                  }}>
                    {g.name?.[0]?.toUpperCase() ?? '?'}
                  </div>

                  <div className="req-card-info">
                    <div className="req-card-name">{g.name}</div>
                    {g.description && (
                      <div style={{ fontSize: '13px', color: 'var(--muted)', marginTop: '2px' }}>
                        {g.description}
                      </div>
                    )}
                    <div style={{ display: 'flex', gap: '10px', marginTop: '6px', flexWrap: 'wrap' }}>
                      <span style={{ fontSize: '12px', color: 'var(--muted)' }}>
                        👤 {memberCount} member{memberCount !== 1 ? 's' : ''}
                      </span>
                      {courseCount > 0 && (
                        <span style={{ fontSize: '12px', color: 'var(--muted)' }}>
                          📚 {courseCount} course{courseCount !== 1 ? 's' : ''}
                        </span>
                      )}
                    </div>
                  </div>

                  <div style={{ color: 'var(--muted)', fontSize: '18px', flexShrink: 0 }}>›</div>
                </div>
              </div>
            </Link>
          )
        })}
      </div>
    </div>
  )
}

export default Groups
