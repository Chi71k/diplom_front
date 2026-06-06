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
  apiGetUserById,
  apiListCourses,
  apiGetFriends,
} from '../../api'
import { avatarColor } from '../../utils/avatar'

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

  const [group, setGroup]             = useState(null)
  const [suggestions, setSuggestions] = useState([])
  const [memberProfiles, setMemberProfiles] = useState({}) // userId → profile
  const [courseMap, setCourseMap]     = useState({})       // courseId → course
  const [friends, setFriends]         = useState([])
  const [loading, setLoading]         = useState(true)
  const [inviteId, setInviteId]       = useState('')
  const [acting, setActing]           = useState(null)

  const load = async () => {
    setLoading(true)
    try {
      const [g, s, fr] = await Promise.all([
        apiGetGroup(id),
        apiGetGroupSuggestions(id, 5).catch(() => ({ items: [] })),
        apiGetFriends().catch(() => ({ items: [] })),
      ])
      const friendIds = fr.items ?? fr ?? []
      if (friendIds.length > 0) {
        const frResults = await Promise.allSettled(friendIds.map(fid => apiGetUserById(fid)))
        setFriends(friendIds.map((fid, i) => {
          const p = frResults[i].status === 'fulfilled' ? frResults[i].value : null
          return { id: fid, firstName: p?.firstName ?? '', lastName: p?.lastName ?? '', avatarUrl: p?.avatarUrl }
        }))
      } else {
        setFriends([])
      }
      setGroup(g)
      setSuggestions(s.items ?? [])

      if (g.members?.length) {
        const results = await Promise.allSettled(g.members.map((m) => apiGetUserById(m.userId)))
        const map = {}
        g.members.forEach((m, i) => {
          if (results[i].status === 'fulfilled') map[m.userId] = results[i].value
        })
        setMemberProfiles(map)
      }

      if (g.courseIds?.length) {
        const allCourses = await apiListCourses({ limit: 200 }).catch(() => [])
        const courses = Array.isArray(allCourses) ? allCourses : []
        const map = {}
        courses.forEach((c) => { map[c.id] = c })
        setCourseMap(map)
      }
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

        <div className="requests-head" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
              <div style={{
                width: '42px', height: '42px', borderRadius: '10px',
                background: 'var(--blue-50)', display: 'flex',
                alignItems: 'center', justifyContent: 'center',
                fontSize: '20px', flexShrink: 0,
              }}>
                {group.name?.[0]?.toUpperCase() ?? 'G'}
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
                  <span key={cid} className="chip chip-course">
                    {courseMap[cid]?.title ?? cid.slice(0, 8) + '…'}
                  </span>
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

        <div style={{ padding: '0 20px 20px' }}>
          <div style={{ fontWeight: 700, fontSize: '14px', marginBottom: '12px', color: 'var(--text)' }}>
            Members ({memberCount})
          </div>

          {group.members?.map((m) => {
            const isMe      = m.userId === profile?.id
            const p         = memberProfiles[m.userId]
            const name      = p ? `${p.firstName ?? ''} ${p.lastName ?? ''}`.trim() : null
            const initial   = (p?.firstName?.[0] || m.userId?.[0] || '?').toUpperCase()
            const roleStyle = roleColors[m.role] ?? roleColors.member

            return (
              <div
                key={m.userId}
                style={{
                  display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                  padding: '10px 0', borderBottom: '1px solid var(--border)',
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                  <div
                    className="avatar"
                    style={{ width: '38px', height: '38px', fontSize: '14px', background: avatarColor(p?.firstName || m.userId), flexShrink: 0 }}
                  >
                    {p?.avatarUrl ? <img src={p.avatarUrl} alt="" /> : initial}
                  </div>
                  <div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
                      {isMe ? (
                        <span style={{ fontSize: '14px', fontWeight: 600 }}>You</span>
                      ) : (
                        <Link
                          to={`/users/${m.userId}`}
                          style={{ fontSize: '14px', fontWeight: 600, color: 'var(--text)', textDecoration: 'none' }}
                          onMouseEnter={(e) => e.currentTarget.style.color = 'var(--primary)'}
                          onMouseLeave={(e) => e.currentTarget.style.color = 'var(--text)'}
                        >
                          {name || m.userId.slice(0, 8) + '…'}
                        </Link>
                      )}
                      <span style={{
                        fontSize: '11px', fontWeight: 600, padding: '2px 7px', borderRadius: '5px',
                        background: roleStyle.bg, color: roleStyle.color,
                      }}>
                        {m.role}
                      </span>
                    </div>
                    {p?.bio && (
                      <div style={{ fontSize: '12px', color: 'var(--muted)', marginTop: '2px' }}>
                        {p.bio.length > 60 ? p.bio.slice(0, 60) + '…' : p.bio}
                      </div>
                    )}
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
            <div style={{ marginTop: '16px' }}>
              {(() => {
                const memberIds = new Set(group.members?.map(m => m.userId) ?? [])
                const available = friends.filter(f => !memberIds.has(f.id))
                if (available.length === 0) return null
                return (
                  <div style={{ marginBottom: '12px' }}>
                    <div style={{ fontSize: '12px', color: 'var(--muted)', fontWeight: 600, marginBottom: '8px' }}>
                      Invite from partners
                    </div>
                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: '6px' }}>
                      {available.map(f => {
                        const fname = `${f.firstName ?? ''} ${f.lastName ?? ''}`.trim() || f.id?.slice(0, 8)
                        return (
                          <button
                            key={f.id}
                            type="button"
                            className="btn btn-secondary btn-sm"
                            disabled={acting === f.id}
                            onClick={async () => {
                              setActing(f.id)
                              try {
                                await apiInviteMember(id, f.id)
                                toast.success(`${fname} invited!`)
                                load()
                              } catch (e) {
                                toast.error(e.error || 'Failed to invite')
                              } finally { setActing(null) }
                            }}
                          >
                            + {fname}
                          </button>
                        )
                      })}
                    </div>
                  </div>
                )
              })()}

              <form onSubmit={handleInvite} style={{ display: 'flex', gap: '8px' }}>
                <input
                  style={{
                    flex: 1, padding: '9px 12px',
                    border: '1px solid var(--border)', borderRadius: '8px',
                    fontSize: '14px', outline: 'none',
                    background: 'var(--card)', color: 'var(--text)',
                  }}
                  placeholder="Or paste User ID…"
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
            </div>
          )}
        </div>

        {suggestions.length > 0 && (
          <div style={{ borderTop: '1px solid var(--border)', padding: '16px 20px' }}>
            <div style={{ fontWeight: 700, fontSize: '14px', marginBottom: '12px' }}>
              AI-suggested members
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
