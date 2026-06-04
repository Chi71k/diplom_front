import { useState, useEffect, useCallback } from 'react'
import { useToast } from '../../context/ToastContext'
import { avatarColor } from '../../utils/avatar'
import {
  apiAdminListUsers,
  apiAdminActivateUser,
  apiAdminDeactivateUser,
} from '../../api'

const PAGE_SIZE = 20

const RoleBadge = ({ role }) => (
  <span style={{
    fontSize: '11px', fontWeight: 700, padding: '2px 8px',
    borderRadius: '5px',
    background: role === 'admin' ? '#ede9fe' : '#f1f5f9',
    color: role === 'admin' ? '#7c3aed' : '#64748b',
  }}>
    {role ?? 'student'}
  </span>
)

const ActiveBadge = ({ active }) => (
  <span style={{
    fontSize: '11px', fontWeight: 700, padding: '2px 8px',
    borderRadius: '5px',
    background: active ? '#dcfce7' : '#fee2e2',
    color: active ? '#15803d' : '#dc2626',
  }}>
    {active ? 'Active' : 'Inactive'}
  </span>
)

const inputStyle = {
  padding: '9px 12px',
  border: '1px solid var(--border)', borderRadius: '8px',
  fontSize: '14px', outline: 'none', background: '#fff',
}

export default function AdminUsers() {
  const toast = useToast()
  const [users, setUsers]     = useState([])
  const [total, setTotal]     = useState(0)
  const [loading, setLoading] = useState(true)
  const [acting, setActing]   = useState(null)
  const [offset, setOffset]   = useState(0)
  const [search, setSearch]   = useState('')
  const [activeFilter, setActiveFilter] = useState('')

  const load = useCallback(async (searchVal = search, activeVal = activeFilter, off = offset) => {
    setLoading(true)
    try {
      const data = await apiAdminListUsers({
        search: searchVal,
        active: activeVal,
        limit: PAGE_SIZE,
        offset: off,
      })
      setUsers(data.users ?? data.items ?? [])
      setTotal(data.total ?? 0)
    } catch (e) {
      toast.error(e.error || 'Failed to load users')
    } finally {
      setLoading(false)
    }
  }, [search, activeFilter, offset])

  useEffect(() => { load() }, [offset])

  const handleSearch = (e) => {
    e.preventDefault()
    setOffset(0)
    load(search, activeFilter, 0)
  }

  const handleFilterChange = (val) => {
    setActiveFilter(val)
    setOffset(0)
    load(search, val, 0)
  }

  const handleToggleActive = async (user) => {
    setActing(user.id)
    try {
      if (user.isActive) {
        await apiAdminDeactivateUser(user.id)
        toast.success(`${user.firstName} deactivated`)
      } else {
        await apiAdminActivateUser(user.id)
        toast.success(`${user.firstName} activated`)
      }
      setUsers((prev) =>
        prev.map((u) => u.id === user.id ? { ...u, isActive: !u.isActive } : u)
      )
    } catch (e) {
      toast.error(e.error || 'Action failed')
    } finally {
      setActing(null)
    }
  }

  const totalPages = Math.ceil(total / PAGE_SIZE)
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1

  const fmt = (iso) => iso
    ? new Date(iso).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
    : '—'

  return (
    <div className="requests-page">
      <div className="card">

        {/* Header */}
        <div className="requests-head">
          <div>
            <div className="requests-title">User Management</div>
            <div className="requests-sub">
              {total > 0 ? `${total} users total` : 'Manage platform users'}
            </div>
          </div>
        </div>

        {/* Filters */}
        <div style={{ padding: '0 20px 16px', display: 'flex', gap: '10px', flexWrap: 'wrap', alignItems: 'center' }}>
          <form onSubmit={handleSearch} style={{ display: 'flex', gap: '8px', flex: 1, minWidth: '200px' }}>
            <input
              style={{ ...inputStyle, flex: 1 }}
              placeholder="Search by name or email…"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
            <button type="submit" className="btn btn-secondary btn-sm">Search</button>
            {search && (
              <button
                type="button"
                className="btn btn-secondary btn-sm"
                onClick={() => { setSearch(''); setOffset(0); load('', activeFilter, 0) }}
              >
                Clear
              </button>
            )}
          </form>

          <div className="tabs" style={{ gap: '6px' }}>
            {[
              { val: '', label: 'All' },
              { val: 'true', label: 'Active' },
              { val: 'false', label: 'Inactive' },
            ].map(({ val, label }) => (
              <button
                key={val}
                type="button"
                className={`btn btn-sm ${activeFilter === val ? 'btn-primary' : 'btn-secondary'}`}
                onClick={() => handleFilterChange(val)}
              >
                {label}
              </button>
            ))}
          </div>
        </div>

        {/* Table */}
        {loading && <div className="loading-state">Loading users…</div>}

        {!loading && users.length === 0 && (
          <div className="empty-state">No users found.</div>
        )}

        {!loading && users.map((user) => {
          const name = `${user.firstName ?? ''} ${user.lastName ?? ''}`.trim() || user.email
          const initials = (user.firstName?.[0] || user.email?.[0] || '?').toUpperCase()

          return (
            <div key={user.id} className="req-card">
              <div className="req-card-main">
                {/* Avatar */}
                <div
                  className="avatar"
                  style={{
                    width: '40px', height: '40px', fontSize: '15px', flexShrink: 0,
                    background: avatarColor(user.firstName || user.email),
                  }}
                >
                  {initials}
                </div>

                <div className="req-card-info" style={{ flex: 1 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
                    <span className="req-card-name">{name}</span>
                    <RoleBadge role={user.role} />
                    <ActiveBadge active={user.isActive} />
                  </div>
                  <div className="req-card-role">{user.email}</div>
                  <div style={{ fontSize: '12px', color: 'var(--muted)', marginTop: '2px' }}>
                    Joined {fmt(user.createdAt)}
                    {user.id && (
                      <span style={{ marginLeft: '8px', fontFamily: 'monospace', fontSize: '11px', color: 'var(--muted-lt)' }}>
                        {user.id.slice(0, 8)}…
                      </span>
                    )}
                  </div>

                  <div className="req-card-actions">
                    {user.role !== 'admin' && (
                      <button
                        className={`btn btn-sm ${user.isActive ? 'btn-danger' : 'btn-primary'}`}
                        disabled={acting === user.id}
                        onClick={() => handleToggleActive(user)}
                      >
                        {acting === user.id
                          ? '…'
                          : user.isActive ? 'Deactivate' : 'Activate'}
                      </button>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )
        })}

        {/* Pagination */}
        {totalPages > 1 && (
          <div style={{
            display: 'flex', justifyContent: 'center', alignItems: 'center',
            gap: '12px', padding: '16px 20px',
            borderTop: '1px solid var(--border)',
          }}>
            <button
              className="btn btn-secondary btn-sm"
              disabled={offset === 0}
              onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
            >
              ← Prev
            </button>
            <span style={{ fontSize: '13px', color: 'var(--muted)' }}>
              Page {currentPage} of {totalPages}
            </span>
            <button
              className="btn btn-secondary btn-sm"
              disabled={offset + PAGE_SIZE >= total}
              onClick={() => setOffset(offset + PAGE_SIZE)}
            >
              Next →
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
