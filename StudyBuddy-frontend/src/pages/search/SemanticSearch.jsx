import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useToast } from '../../context/ToastContext'
import { apiSearchStudents, apiSendMatchRequest } from '../../api'
import UserAvatar from '../../components/UserAvatar'
import './SemanticSearch.css'

export default function SemanticSearch() {
  const toast = useToast()
  const [query, setQuery]     = useState('')
  const [results, setResults] = useState(null)   // null = not searched yet
  const [searchError, setSearchError] = useState(null)
  const [loading, setLoading] = useState(false)
  const [sending, setSending] = useState(null)

  const handleSearch = async (e) => {
    e.preventDefault()
    if (!query.trim()) return
    setLoading(true)
    setResults(null)
    setSearchError(null)
    try {
      const data = await apiSearchStudents(query.trim(), 20)
      setResults(data.items ?? [])
    } catch (e) {
      const msg = e.status === 500
        ? 'AI search is unavailable — the GEMINI_API_KEY on the server may be missing or invalid.'
        : (e.error || 'Search failed')
      setSearchError(msg)
      toast.error(msg)
    } finally {
      setLoading(false)
    }
  }

  const handleConnect = async (userId) => {
    setSending(userId)
    try {
      await apiSendMatchRequest(userId, '')
      toast.success('Connect request sent!')
    } catch (e) {
      toast.error(e.error || 'Failed to send request')
    } finally {
      setSending(null)
    }
  }

  const matchClass = (score) =>
    score >= 0.7 ? 'match-high' : score >= 0.4 ? 'match-mid' : 'match-low'

  return (
    <div className="find-page">
      <div className="card">
        <div className="ai-hero">
          <div className="ai-hero-icon">
            <svg width="26" height="26" viewBox="0 0 24 24" fill="currentColor">
              <path d="m12 3-1.912 5.813a2 2 0 0 1-1.275 1.275L3 12l5.813 1.912a2 2 0 0 1 1.275 1.275L12 21l1.912-5.813a2 2 0 0 1 1.275-1.275L21 12l-5.813-1.912a2 2 0 0 1-1.275-1.275L12 3Z"/>
            </svg>
          </div>
          <div className="ai-hero-title">AI Partner Search</div>
          <div className="ai-hero-sub">
            Describe what you're looking for in plain language — AI finds semantically similar study partners
          </div>
        </div>

        <form onSubmit={handleSearch} className="sem-search-form">
          <textarea
            className="sem-search-input"
            placeholder={`e.g. "Looking for someone studying machine learning and available in the evenings"`}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            rows={3}
            maxLength={500}
          />
          <div className="sem-search-actions">
            <span className="sem-search-hint">{query.length}/500</span>
            <button
              type="submit"
              className="btn btn-ai"
              disabled={loading || !query.trim()}
            >
              {loading
                ? <><span className="sem-spinner" />Searching…</>
                : <>
                    <svg width="15" height="15" viewBox="0 0 24 24" fill="currentColor">
                      <path d="m12 3-1.912 5.813a2 2 0 0 1-1.275 1.275L3 12l5.813 1.912a2 2 0 0 1 1.275 1.275L12 21l1.912-5.813a2 2 0 0 1 1.275-1.275L21 12l-5.813-1.912a2 2 0 0 1-1.275-1.275L12 3Z"/>
                    </svg>
                    AI Search
                  </>
              }
            </button>
          </div>
        </form>

        {searchError && !loading && (
          <div className="ai-error-box">
            <strong>Search unavailable:</strong> {searchError}
            <div className="ai-error-hint">
              Set a valid <code>GEMINI_API_KEY</code> env variable and restart the users container.
            </div>
          </div>
        )}

        {results !== null && results.length === 0 && !loading && !searchError && (
          <div className="empty-state">No matches found. Try a different description.</div>
        )}

        {results !== null && results.length > 0 && !loading && (
          <div>
            <div className="ai-results-count">
              <span className="ai-results-count-num">{results.length}</span>
              {' '}result{results.length !== 1 ? 's' : ''} — sorted by semantic similarity
            </div>
            <div className="ai-results-list">
              {results.map((row, i) => {
                const name  = `${row.firstName ?? ''} ${row.lastName ?? ''}`.trim() || '—'
                const score = row.similarity ?? 0
                const pct   = Math.round(score * 100)

                return (
                  <div
                    key={row.userId ?? i}
                    className="ai-result-card"
                    style={{ animationDelay: `${Math.min(i, 6) * 55}ms` }}
                  >
                    <Link to={`/users/${row.userId}`} className="cand-avatar-link">
                      <UserAvatar user={row} size="avatar-md" />
                    </Link>
                    <div className="ai-result-body">
                      <div className="ai-result-top">
                        <Link to={`/users/${row.userId}`} className="cand-name-link">{name}</Link>
                        <div className={`ai-match-badge ${matchClass(score)}`}>
                          <span className="ai-match-pct">{pct}%</span>
                          <span className="ai-match-label">match</span>
                        </div>
                      </div>
                      {row.bio && <div className="cand-bio">{row.bio}</div>}
                      <div className="cand-actions">
                        <button
                          type="button"
                          className="btn btn-primary btn-sm"
                          disabled={sending === row.userId}
                          onClick={() => handleConnect(row.userId)}
                        >
                          {sending === row.userId ? 'Sending…' : 'Connect'}
                        </button>
                        <Link to={`/users/${row.userId}`} className="btn btn-secondary btn-sm">
                          View profile
                        </Link>
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
