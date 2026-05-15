import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useToast } from '../../context/ToastContext'
import { apiCreateGroup, apiGetGroup, apiUpdateGroup } from '../../api'

const inputStyle = {
  width: '100%', padding: '10px 13px',
  border: '1px solid var(--border)', borderRadius: '8px',
  fontSize: '14px', outline: 'none', background: '#fff',
}

const GroupForm = ({ edit }) => {
  const toast    = useToast()
  const navigate = useNavigate()
  const { id }   = useParams()
  const [form, setForm]     = useState({ name: '', description: '', courseIds: '' })
  const [loading, setLoading] = useState(!!edit)
  const [saving, setSaving]   = useState(false)

  useEffect(() => {
    if (!edit) return
    const load = async () => {
      try {
        const g = await apiGetGroup(id)
        setForm({
          name: g.name,
          description: g.description ?? '',
          courseIds: (g.courseIds ?? []).join(', '),
        })
      } catch {
        toast.error('Failed to load group')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [id])

  const handleSubmit = async (e) => {
    e.preventDefault()
    setSaving(true)
    const body = {
      name: form.name,
      description: form.description,
      courseIds: form.courseIds.split(',').map((s) => s.trim()).filter(Boolean),
    }
    try {
      if (edit) {
        await apiUpdateGroup(id, body)
        toast.success('Group updated!')
        navigate(`/groups/${id}`)
      } else {
        const g = await apiCreateGroup(body)
        toast.success('Group created!')
        navigate(`/groups/${g.id}`)
      }
    } catch (e) {
      toast.error(e.error || 'Failed to save group')
    } finally {
      setSaving(false)
    }
  }

  if (loading) return <div className="loading-state">Loading…</div>

  return (
    <div className="requests-page">
      <div className="card" style={{ padding: '24px' }}>
        <div style={{ marginBottom: '24px' }}>
          <div className="requests-title">{edit ? 'Edit Group' : 'Create Group'}</div>
          <div className="requests-sub">{edit ? 'Update group details' : 'Start a new study group'}</div>
        </div>

        <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          <div>
            <label style={{ fontSize: '13px', fontWeight: 600, color: '#374151', marginBottom: '6px', display: 'block' }}>
              Group name *
            </label>
            <input
              style={inputStyle}
              required
              value={form.name}
              onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              placeholder="e.g. Algorithms Study Group"
            />
          </div>

          <div>
            <label style={{ fontSize: '13px', fontWeight: 600, color: '#374151', marginBottom: '6px', display: 'block' }}>
              Description
              <span style={{ fontWeight: 400, color: 'var(--muted)', marginLeft: '6px' }}>(optional)</span>
            </label>
            <textarea
              style={{ ...inputStyle, minHeight: '90px', resize: 'vertical' }}
              rows={3}
              value={form.description}
              onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
              placeholder="What will this group focus on?"
            />
          </div>

          <div>
            <label style={{ fontSize: '13px', fontWeight: 600, color: '#374151', marginBottom: '6px', display: 'block' }}>
              Course IDs
              <span style={{ fontWeight: 400, color: 'var(--muted)', marginLeft: '6px' }}>(comma-separated, optional)</span>
            </label>
            <input
              style={inputStyle}
              value={form.courseIds}
              onChange={(e) => setForm((f) => ({ ...f, courseIds: e.target.value }))}
              placeholder="Paste course UUIDs separated by commas"
            />
          </div>

          <div style={{ display: 'flex', gap: '10px', paddingTop: '4px' }}>
            <button type="submit" className="btn btn-primary" disabled={saving}>
              {saving ? 'Saving…' : edit ? 'Save changes' : 'Create group'}
            </button>
            <button
              type="button"
              className="btn btn-secondary"
              onClick={() => navigate(edit ? `/groups/${id}` : '/groups')}
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default GroupForm
