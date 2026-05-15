import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { useToast } from '../../context/ToastContext'
import { useAuth } from '../../context/useAuth'
import {
  apiGetGroup,
  apiDeleteGroup,
  apiInviteMember,
  apiRemoveMember,
  apiGetGroupSuggestions,
} from '../../api'
import { avatarColor } from '../../utils/avatar'

const shortId = (id) => id?.slice(0, 8) ?? '?'

const roleColors = {
  owner:  { bg: '#eff6ff', color: '#2563eb' },
  admin:  { bg: '#f0fdf4', color: '#15803d' },
  member: { bg: '#f8fafc', color: '#64748b' },
}

const GroupDetail = () => {
  const { id }  = useParams()
  const navigate = useNavigate()
  const toast    = useToast()
  const { profile } = useAuth()

  const [group, setGroup]           = useState(null)
  const [suggestions, setSuggestions] = useState([])
  const [loading, setLoading]       = useState(true)
  const [inviteId, setInviteId]     = useState('')
  const [acting, setActing]         = useState(null)

  const load = async () => {
    setLoading(true)
    try {
      const [g, s] = await Promise.all([
        apiGetGroup(id),
        apiGetGroupSuggestions(id, 5).catch(() => ({ items: [] })),
      ])
      setGroup(g)
      setSuggestions(s.items ?? [])
    } catch (e) {
      toast.error(e.error || 'Failed to load group')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [id])

  const isOwner = group?.ownerId === profile?.id

  const handleDelete = async () => {
    if (!confirm('Delete this group? This cannot be undone.')) return
    try {
      await apiDeleteGroup(id)
      toast.success('Group deleted')
      navigate('/groups')
    } catch (e) {
      toast.error(e.error || 'Failed to delete')
    }
  }

  const handleInvite = async (e) => {
    e.preventDefault()
    if (!inviteId.trim()) return
    setActing('invite')
    try {
      await apiInviteMember(id, inviteId.trim())
      toast.success('Member invited!')
      setInviteId('')
      load()
    } catch (e) {
      toast.error(e.error || 'Failed to invite')
    } finally {
      setActing(null)
    }
  }

  const handleRemove = async (userId) => {
    setActing(userId)
    try {
      await apiRemoveMember(id, userId)
      toast.success('Member removed')
      setGroup((g) => ({ ...g, members: g.members.filter((m) => m.userId !== userId) }))
    } catch (e) {
      toast.error(e.error || 'Failed to remove')
    } finally {
      setActing(null)
    }
  }

  const handleSuggestInvite = async (userId) => {
    setActing(userId)
    try {
      await apiInviteMember(id, userId)
      toast.success('Invited!')
      setSuggestions((prev) => prev.filter((s) => s.userId !== userId))
      load()
    } catch (e) {
      toast.error(e.error || 'Failed to invite')
    } finally {
      setActing(null)
    }
  }

  if (loading) return <div className="loading-state">Loading…</div>
  if (!group)  return <div className="empty-state">Group not found.</div>

  const memberCount = group.members?.length ?? 0

  return (
    <div className="requests-page">
      <div className="card">

        {/* Header */}
        <div className="requests-head" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
              <div style={{
                width: '42px', height: '42px', borderRadius: '10px',
                background: 'var(--blue-50)', display: 'flex',
                alignItems: 'center', justifyContent: 'center',
                fontSize: '20px', flexShrink: 0,
              }}>
                👥
              </div>
              <div>
                <div className="requests-title" style={{ marginBottom: 0 }}>{group.name}</div>
                {group.description && (
                  <div className="requests-sub">{group.description}</div>
                )}
              </div>
            </div>
            {group.courseIds?.length > 0 && (
              <div style={{ marginTop: '10px', display: 'flex', flexWrap: 'wrap', gap: '6px', paddingLeft: '52px' }}>
                {group.courseIds.map((cid) => (
                  <span key={cid} className="chip chip-course">{cid.slice(0, 8)}</span>
                ))}
              </div>
            )}
          </div>

          {isOwner && (
            <div style={{ display: 'flex', gap: '8px', flexShrink: 0 }}>
              <Link to={`/groups/${id}/edit`} className="btn btn-secondary btn-sm">Edit</Link>
              <button className="btn btn-danger btn-sm" onClick={handleDelete}>Delete</button>
            </div>
          )}
        </div>

        {/* Members */}
        <div style={{ padding: '0 20px 20px' }}>
          <div style={{ fontWeight: 700, fontSize: '14px', marginBottom: '12px', color: 'var(--text)' }}>
            Members ({memberCount})
          </div>

          {group.members?.map((m) => {
            const isMe = m.userId === profile?.id
            const initials = m.userId?.[0]?.toUpperCase() ?? '?'
            const roleStyle = roleColors[m.role] ?? roleColors.member
            return (
              <div
                key={m.userId}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  padding: '10px 0',
                  borderBottom: '1px solid var(--border)',
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                  <div
                    className="avatar"
                    style={{ width: '36px', height: '36px', fontSize: '14px', background: avatarColor(m.userId) }}
                  >
                    {initials}
                  </div>
                  <div>
                    <span style={{ fontSize: '14px', fontWeight: 500 }}>
                      {isMe ? 'You' : shortId(m.userId)}
                      {isMe && (
                        <span style={{ fontSize: '11px', color: 'var(--muted)', marginLeft: '6px' }}>
                          ({shortId(m.userId)}…)
                        </span>
                      )}
                    </span>
                    <span
                      style={{
                        marginLeft: '8px',
                        fontSize: '11px',
                        fontWeight: 600,
                        padding: '2px 7px',
                        borderRadius: '5px',
                        background: roleStyle.bg,
                        color: roleStyle.color,
                      }}
                    >
                      {m.role}
                    </span>
                  </div>
                </div>

                {isOwner && !isMe && (
                  <button
                    className="btn btn-danger btn-sm"
                    disabled={acting === m.userId}
                    onClick={() => handleRemove(m.userId)}
                  >
                    Remove
                  </button>
                )}
              </div>
            )
          })}

          {/* Invite */}
          {isOwner && (
            <form onSubmit={handleInvite} style={{ display: 'flex', gap: '8px', marginTop: '16px' }}>
              <input
                style={{
                  flex: 1, padding: '9px 12px',
                  border: '1px solid var(--border)', borderRadius: '8px',
                  fontSize: '14px', outline: 'none',
                }}
                placeholder="Paste User ID to invite…"
                value={inviteId}
                onChange={(e) => setInviteId(e.target.value)}
              />
              <button
                type="submit"
                className="btn btn-primary btn-sm"
                disabled={acting === 'invite'}
              >
                Invite
              </button>
            </form>
          )}
        </div>

        {/* AI suggestions */}
        {suggestions.length > 0 && (
          <div style={{ borderTop: '1px solid var(--border)', padding: '16px 20px' }}>
            <div style={{ fontWeight: 700, fontSize: '14px', marginBottom: '12px', display: 'flex', alignItems: 'center', gap: '6px' }}>
              <span>✨</span> AI-suggested members
            </div>
            {suggestions.map((s) => (
              <div key={s.userId} className="cand-card">
                <div className="cand-main">
                  <div
                    className="avatar avatar-sm"
                    style={{ background: avatarColor(s.firstName) }}
                  >
                    {s.avatarUrl
                      ? <img src={s.avatarUrl} alt="" />
                      : (s.firstName?.[0] || '?').toUpperCase()
                    }
                  </div>
                  <div className="cand-info">
                    <div className="cand-name">{s.firstName} {s.lastName}</div>
                    <div style={{ fontSize: '12px', color: 'var(--muted)' }}>
                      {Math.round(s.similarityScore * 100)}% match
                    </div>
                    {isOwner && (
                      <button
                        className="btn btn-primary btn-sm"
                        style={{ marginTop: '8px' }}
                        disabled={acting === s.userId}
                        onClick={() => handleSuggestInvite(s.userId)}
                      >
                        Invite
                      </button>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

export default GroupDetail
