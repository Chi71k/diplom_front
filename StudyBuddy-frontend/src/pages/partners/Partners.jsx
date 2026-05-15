import { useState, useEffect } from 'react'
import { useAuth } from '../../context/useAuth'
import { useToast } from '../../context/ToastContext'
import { apiGetMatchRequests, apiGetUserById } from '../../api'
import { Link } from 'react-router-dom'
import { avatarColor } from '../../utils/avatar'

const Partners = () => {
  const { profile } = useAuth()
  const toast = useToast()
  const [partners, setPartners] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const load = async () => {
      setLoading(true)
      try {
        const data = await apiGetMatchRequests({ status: 'accepted', limit: 100 })
        const items = data.items || []
        const myId = profile?.id
        const partnerIds = items.map((r) => r.requesterId === myId ? r.receiverId : r.requesterId)

        const results = await Promise.allSettled(partnerIds.map((id) => apiGetUserById(id)))

        const enriched = items
          .map((request, i) => {
            const partnerId = partnerIds[i]
            const partner = results[i].status === 'fulfilled'
              ? results[i].value
              : { id: partnerId, firstName: partnerId?.slice(0, 8), lastName: '' }
            return { request, partner }
          })

        setPartners(enriched)
      } catch (e) {
        toast.error(e.error || 'Failed to load partners')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [profile?.id])

  const fmt = (iso) =>
    new Date(iso).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })

  return (
    <div className="requests-page">
      <div className="card">
        <div className="requests-head">
          <div className="requests-title">My Partners</div>
          <div className="requests-sub">
            {partners.length > 0
              ? `You have ${partners.length} study partner${partners.length > 1 ? 's' : ''}.`
              : 'Accepted connections will appear here.'
            }
          </div>
        </div>

        {loading && <div className="loading-state">Loading partners...</div>}

        {!loading && partners.length === 0 && (
          <div className="empty-state">
            No partners yet.{' '}
            <Link to="/matching/candidates" style={{ color: 'var(--primary)', fontWeight: 600 }}>
              Find partners
            </Link>{' '}
            to send a request.
          </div>
        )}

        {!loading && partners.map(({ request, partner }) => (
          <div key={request.id} className="req-card">
            <div className="req-card-main">
              <div className="avatar avatar-md" style={{ background: avatarColor(partner.firstName) }}>
                {partner.avatarUrl
                  ? <img src={partner.avatarUrl} alt="" />
                  : (partner.firstName?.[0] || '?').toUpperCase()
                }
              </div>

              <div className="req-card-info">
                <div className="req-card-name">{partner.firstName} {partner.lastName}</div>
                <div className="req-card-role">
                  Partners since {fmt(request.updatedAt ?? request.createdAt)}
                </div>
                {partner.bio && (
                  <div style={{ fontSize: '13px', color: 'var(--muted)', marginTop: '6px', lineHeight: 1.5 }}>
                    {partner.bio}
                  </div>
                )}
                {request.message && (
                  <div className="req-card-msg">"{request.message}"</div>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default Partners
