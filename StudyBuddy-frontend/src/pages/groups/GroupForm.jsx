import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useToast } from '../../context/ToastContext'
import { apiCreateGroup, apiGetGroup, apiUpdateGroup, apiListCourses } from '../../api'
import { useAuth } from '../../context/useAuth'

const inputStyle = {
  width: '100%', padding: '10px 13px',
  border: '1px solid var(--border)', borderRadius: '8px',
  fontSize: '14px', outline: 'none', background: '#fff',
}

const labelStyle = {
  fontSize: '13px', fontWeight: 600, color: '#374151',
  marginBottom: '6px', display: 'block',
}

const GroupForm = ({ edit }) => {
  const toast        = useToast()
  const navigate     = useNavigate()
  const { id }       = useParams()
  const { profile }  = useAuth()

  const [form, setForm]           = useState({ name: '', description: '' })
  const [selectedCourseIds, setSelectedCourseIds] = useState([])
  const [courses, setCourses]     = useState([])
  const [loading, setLoading]     = useState(true)
  const [saving, setSaving]       = useState(false)

  useEffect(() => {
    const load = async () => {
      setLoading(true)
      try {
        const coursesData = await apiListCourses({ limit: 100 })
        const all = Array.isArray(coursesData) ? coursesData : []
        setCourses(all.filter((c) => c.ownerUserId === profile?.id))

        if (edit) {
          const g = await apiGetGroup(id)
          setForm({ name: g.name, description: g.description ?? '' })
          setSelectedCourseIds(g.courseIds ?? [])
        }
      } catch {
        toast.error('Failed to load data')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [id])

  const toggleCourse = (courseId) => {
    setSelectedCourseIds((prev) =>
      prev.includes(courseId)
        ? prev.filter((c) => c !== courseId)
        : [...prev, courseId]
    )
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setSaving(true)
    const body = {
      name:        form.name,
      description: form.description,
      courseIds:   selectedCourseIds,
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
            <label style={labelStyle}>Group name *</label>
            <input
              style={inputStyle}
              required
              value={form.name}
              onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              placeholder="e.g. Algorithms Study Group"
            />
          </div>

          <div>
            <label style={labelStyle}>
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
            <label style={labelStyle}>
              Courses
              <span style={{ fontWeight: 400, color: 'var(--muted)', marginLeft: '6px' }}>(optional)</span>
            </label>

            {courses.length === 0 ? (
              <div style={{ fontSize: '13px', color: 'var(--muted)', padding: '10px 0' }}>
                No courses yet.{' '}
                <span
                  style={{ color: 'var(--primary)', cursor: 'pointer', fontWeight: 600 }}
                  onClick={() => navigate('/courses/new')}
                >
                  Add a course first
                </span>
              </div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                {courses.map((course) => {
                  const selected = selectedCourseIds.includes(course.id)
                  return (
                    <label
                      key={course.id}
                      style={{
                        display: 'flex', alignItems: 'center', gap: '10px',
                        padding: '10px 12px', borderRadius: '8px', cursor: 'pointer',
                        border: `1.5px solid ${selected ? 'var(--primary)' : 'var(--border)'}`,
                        background: selected ? 'var(--blue-50)' : '#fff',
                        transition: 'border-color .15s, background .15s',
                      }}
                    >
                      <input
                        type="checkbox"
                        checked={selected}
                        onChange={() => toggleCourse(course.id)}
                        style={{ width: '16px', height: '16px', accentColor: 'var(--primary)', flexShrink: 0 }}
                      />
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{ fontSize: '14px', fontWeight: selected ? 600 : 400, color: 'var(--text)' }}>
                          {course.title}
                        </div>
                        {(course.subject || course.level) && (
                          <div style={{ fontSize: '12px', color: 'var(--muted)', marginTop: '1px' }}>
                            {[course.subject, course.level].filter(Boolean).join(' · ')}
                          </div>
                        )}
                      </div>
                      {selected && (
                        <span style={{ fontSize: '12px', color: 'var(--primary)', fontWeight: 700, flexShrink: 0 }}>
                          ✓
                        </span>
                      )}
                    </label>
                  )
                })}
              </div>
            )}

            {selectedCourseIds.length > 0 && (
              <div style={{ fontSize: '12px', color: 'var(--muted)', marginTop: '6px' }}>
                {selectedCourseIds.length} course{selectedCourseIds.length > 1 ? 's' : ''} selected
              </div>
            )}
          </div>

          <div style={{ display: 'flex', gap: '10px', paddingTop: '4px' }}>
            <button type="submit" className="btn btn-primary" disabled={saving}>
              {saving ? 'Saving…' : edit ? 'Save changes' : 'Create group'}
            </button>
            <button
              type="button"
              className="btn btn-secondary"
              onClick={() => navigate(edit ? `/groups/${id}` : '/study')}
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
