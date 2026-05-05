import { useState, useEffect } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useAuth } from '../../context/useAuth'
import { useToast } from '../../context/ToastContext'
import { apiGetProfile, apiUpdateProfile, apiDeleteProfile, apiGetMyInterests } from '../../api'

const Profile = () => {
  const navigate = useNavigate()
  const { profile, setProfile, setToken } = useAuth()
  const toast = useToast()
  const [loading, setLoading] = useState(!profile)
  const [loadError, setLoadError] = useState('')
  const [editing, setEditing] = useState(false)
  const [form, setForm] = useState({ firstName: '', lastName: '', bio: '', avatarUrl: '' })
  const [saving, setSaving] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState(false)
  const [avatarError, setAvatarError] = useState(false)
  const [interests, setInterests] = useState([])

  const load = async () => {
    setLoadError('')
    setLoading(true)
    try {
      const [data, interestsData] = await Promise.all([apiGetProfile(), apiGetMyInterests()])
      setProfile(data)
      setInterests(interestsData.items ?? [])
      setForm({
        firstName: data.firstName || '',
        lastName: data.lastName || '',
        bio: data.bio || '',
        avatarUrl: data.avatarUrl || '',
      })
    } catch (e) {
      setLoadError(e.error || 'Failed to load profile')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (profile) {
      setForm({
        firstName: profile.firstName || '',
        lastName: profile.lastName || '',
        bio: profile.bio || '',
        avatarUrl: profile.avatarUrl || '',
      })
      apiGetMyInterests()
        .then((d) => setInterests(d.items ?? []))
        .catch(() => {})
    } else {
      load()
    }
  }, [])

  useEffect(() => { setAvatarError(false) }, [profile?.avatarUrl])

  const handleSave = async (e) => {
    e.preventDefault()
    setSaving(true)
    try {
      const body = {}
      if (form.firstName !== (profile?.firstName ?? '')) body.firstName = form.firstName
      if (form.lastName  !== (profile?.lastName  ?? '')) body.lastName  = form.lastName
      if (form.bio       !== (profile?.bio       ?? '')) body.bio       = form.bio
      if (form.avatarUrl !== (profile?.avatarUrl ?? '')) body.avatarUrl = form.avatarUrl
      const data = await apiUpdateProfile(body)
      setProfile(data)
      setEditing(false)
      toast.success('Profile saved')
    } catch (err) {
      toast.error(err.error || 'Failed to save')
    } finally {
      setSaving(false)
    }
  }

  const handleDeleteAccount = async () => {
    if (!deleteConfirm) { setDeleteConfirm(true); return }
    setSaving(true)
    try {
      await apiDeleteProfile()
      setProfile(null)
      setToken(null)
      navigate('/login')
    } catch (err) {
      toast.error(err.error || 'Failed to delete account')
    } finally {
      setSaving(false)
    }
  }

  const initial = (profile?.firstName?.[0] || profile?.email?.[0] || '?').toUpperCase()
  const showAvatar = profile?.avatarUrl && !avatarError

  if (loading) return <div className="loading-state">Loading profile...</div>

  if (loadError && !profile) {
    return (
      <div className="profile-page">
        <div className="card" style={{ padding: '24px', textAlign: 'center' }}>
          <div className="auth-error" style={{ marginBottom: '16px' }}>{loadError}</div>
          <button onClick={load} className="btn btn-primary">Try again</button>
        </div>
      </div>
    )
  }

  if (!profile) return <div className="empty-state">You are not signed in.</div>

  return (
    <div className="profile-page">
      {/* Cover + head */}
      <div className="profile-cover" />
      <div className="card profile-head-card">
        <div className="profile-ava-row">
          <div className="profile-ava">
            {showAvatar
              ? <img src={profile.avatarUrl} alt="" onError={() => setAvatarError(true)} />
              : initial
            }
          </div>
          <button
            type="button"
            className="btn btn-secondary btn-sm"
            onClick={() => setEditing(true)}
          >
            Edit profile
          </button>
        </div>
        <div className="profile-name">{profile.firstName} {profile.lastName}</div>
        <div className="profile-subtitle">{profile.email}</div>
        {profile.bio && <div className="profile-loc" style={{ marginTop: '8px', fontSize: '13px', color: 'var(--muted)', lineHeight: 1.5 }}>{profile.bio}</div>}

        <div className="profile-stats-row">
          <div className="profile-stat">
            <div className="profile-stat-val">{interests.length}</div>
            <div className="profile-stat-lbl">Interests</div>
          </div>
          <div className="profile-stat">
            <div className="profile-stat-val">
              <Link to="/courses" style={{ color: 'inherit' }}>Courses</Link>
            </div>
            <div className="profile-stat-lbl"><Link to="/courses" style={{ color: 'var(--muted)' }}>View</Link></div>
          </div>
          <div className="profile-stat">
            <div className="profile-stat-val">
              <Link to="/matching/partners" style={{ color: 'inherit' }}>Partners</Link>
            </div>
            <div className="profile-stat-lbl"><Link to="/matching/partners" style={{ color: 'var(--muted)' }}>View</Link></div>
          </div>
        </div>
      </div>

      {/* Edit form */}
      {editing && (
        <div className="card p-section" style={{ marginTop: '14px' }}>
          <div className="p-section-head">
            <span className="p-section-title">Edit profile</span>
            <button type="button" className="btn btn-ghost btn-sm" onClick={() => setEditing(false)}>Cancel</button>
          </div>
          <form onSubmit={handleSave} className="profile-form p-section-body">
            <label className="profile-label">First name</label>
            <input className="profile-input" value={form.firstName}
              onChange={(e) => setForm((f) => ({ ...f, firstName: e.target.value }))}
              placeholder="First name" required />
            <label className="profile-label">Last name</label>
            <input className="profile-input" value={form.lastName}
              onChange={(e) => setForm((f) => ({ ...f, lastName: e.target.value }))}
              placeholder="Last name" required />
            <label className="profile-label">Bio</label>
            <textarea className="profile-input profile-textarea" value={form.bio}
              onChange={(e) => setForm((f) => ({ ...f, bio: e.target.value }))}
              placeholder="Tell us about your study goals" rows={3} />
            <label className="profile-label">Profile photo URL</label>
            <input className="profile-input" type="url" value={form.avatarUrl}
              onChange={(e) => setForm((f) => ({ ...f, avatarUrl: e.target.value }))}
              placeholder="https://example.com/photo.jpg" />
            <p className="profile-field-hint">Paste a direct link to an image file.</p>
            <div className="profile-form-actions">
              <button type="submit" className="btn btn-primary" disabled={saving}>
                {saving ? 'Saving...' : 'Save changes'}
              </button>
              <button type="button" className="btn btn-secondary" onClick={() => setEditing(false)}>
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Interests section */}
      {interests.length > 0 && (
        <div className="card p-section" style={{ marginTop: '14px' }}>
          <div className="p-section-head">
            <span className="p-section-title">Interests</span>
            <Link to="/interests" className="btn btn-ghost btn-sm">Edit</Link>
          </div>
          <div className="p-section-body">
            <div className="chips-row">
              {interests.map((i) => (
                <span key={i.ID} className="chip chip-int">{i.Name}</span>
              ))}
            </div>
          </div>
        </div>
      )}

      {interests.length === 0 && (
        <div className="card p-section" style={{ marginTop: '14px' }}>
          <div className="p-section-head">
            <span className="p-section-title">Interests</span>
          </div>
          <div className="p-section-body">
            <p className="page-muted" style={{ fontSize: '13px' }}>
              No interests added yet.{' '}
              <Link to="/interests" style={{ color: 'var(--primary)' }}>Add interests</Link> to improve your match score.
            </p>
          </div>
        </div>
      )}

      {/* Danger zone */}
      <div className="card p-section profile-danger-card" style={{ marginTop: '14px' }}>
        <div className="p-section-head">
          <span className="p-section-title">Danger zone</span>
        </div>
        <div className="p-section-body">
          <p className="profile-danger-text">Once deleted, your data cannot be recovered.</p>
          <div style={{ display: 'flex', gap: '10px', flexWrap: 'wrap' }}>
            <button type="button" className="btn btn-danger" onClick={handleDeleteAccount} disabled={saving}>
              {deleteConfirm ? 'Click again to confirm' : 'Delete account'}
            </button>
            {deleteConfirm && (
              <button type="button" className="btn btn-ghost" onClick={() => setDeleteConfirm(false)}>
                Cancel
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default Profile
