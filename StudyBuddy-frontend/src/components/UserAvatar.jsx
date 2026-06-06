import { useState } from 'react'
import { avatarColor } from '../utils/avatar'

export default function UserAvatar({ user, size = 'avatar-md' }) {
  const [imgErr, setImgErr] = useState(false)
  const name   = user?.firstName || user?.email || ''
  const letter = (name[0] || '?').toUpperCase()
  const bg     = avatarColor(name || user?.id || '')

  return (
    <div className={`avatar ${size}`} style={{ background: bg }}>
      {user?.avatarUrl && !imgErr
        ? <img src={user.avatarUrl} alt="" onError={() => setImgErr(true)} />
        : letter
      }
    </div>
  )
}
