import { useState, useEffect } from 'react'
import { useToast } from '../../context/ToastContext'
import { apiAdminStats } from '../../api'

const StatCard = ({ label, value, color = 'var(--primary)' }) => (
  <div style={{
    padding: '20px 24px',
    borderRadius: '12px',
    background: '#fff',
    border: '1px solid var(--border)',
    textAlign: 'center',
    flex: '1 1 140px',
  }}>
    <div style={{ fontSize: '36px', fontWeight: 800, color, lineHeight: 1 }}>
      {value ?? '—'}
    </div>
    <div style={{ fontSize: '13px', color: 'var(--muted)', marginTop: '6px' }}>{label}</div>
  </div>
)

export default function AdminStats() {
  const toast = useToast()
  const [stats, setStats]     = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const load = async () => {
      setLoading(true)
      try {
        const data = await apiAdminStats()
        setStats(data)
      } catch (e) {
        toast.error(e.error || 'Failed to load stats')
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
        <div className="requests-head">
          <div>
            <div className="requests-title">Platform Statistics</div>
            <div className="requests-sub">Overview of platform activity</div>
          </div>
        </div>

        {loading && <div className="loading-state">Loading stats…</div>}

        {!loading && !stats && (
          <div className="empty-state">No statistics available.</div>
        )}

        {!loading && stats && (
          <div style={{ padding: '8px 20px 24px' }}>

            {/* Users block */}
            <div style={{ fontSize: '12px', fontWeight: 700, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: '0.5px', marginBottom: '12px' }}>
              Users
            </div>
            <div style={{ display: 'flex', gap: '12px', flexWrap: 'wrap', marginBottom: '24px' }}>
              <StatCard label="Total users"    value={stats.totalUsers}   color="var(--primary)" />
              <StatCard label="Active users"   value={stats.activeUsers}  color="#15803d" />
              <StatCard label="Inactive users" value={stats.inactiveUsers ?? ((stats.totalUsers ?? 0) - (stats.activeUsers ?? 0))} color="#dc2626" />
            </div>

            {/* Activity block */}
            <div style={{ fontSize: '12px', fontWeight: 700, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: '0.5px', marginBottom: '12px' }}>
              Activity
            </div>
            <div style={{ display: 'flex', gap: '12px', flexWrap: 'wrap', marginBottom: '24px' }}>
              {stats.totalMatches   !== undefined && <StatCard label="Matches"   value={stats.totalMatches}   color="#d97706" />}
              {stats.totalSessions  !== undefined && <StatCard label="Sessions"  value={stats.totalSessions}  color="#7c3aed" />}
              {stats.totalGroups    !== undefined && <StatCard label="Groups"    value={stats.totalGroups}    color="#0891b2" />}
              {stats.totalCourses   !== undefined && <StatCard label="Courses"   value={stats.totalCourses}   color="#15803d" />}
              {stats.totalReviews   !== undefined && <StatCard label="Reviews"   value={stats.totalReviews}   color="#be185d" />}
            </div>

            {/* Raw dump for any extra fields */}
            {Object.keys(stats).some((k) => ![
              'totalUsers', 'activeUsers', 'inactiveUsers',
              'totalMatches', 'totalSessions', 'totalGroups', 'totalCourses', 'totalReviews',
            ].includes(k)) && (
              <div>
                <div style={{ fontSize: '12px', fontWeight: 700, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: '0.5px', marginBottom: '12px' }}>
                  Other
                </div>
                <div style={{ display: 'flex', gap: '12px', flexWrap: 'wrap' }}>
                  {Object.entries(stats)
                    .filter(([k]) => ![
                      'totalUsers', 'activeUsers', 'inactiveUsers',
                      'totalMatches', 'totalSessions', 'totalGroups', 'totalCourses', 'totalReviews',
                    ].includes(k))
                    .map(([k, v]) => (
                      <StatCard key={k} label={k} value={typeof v === 'object' ? JSON.stringify(v) : String(v)} />
                    ))
                  }
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
