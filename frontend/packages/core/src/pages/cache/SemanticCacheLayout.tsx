import { useEffect } from 'react'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'

export default function SemanticCacheLayout() {
  const location = useLocation()
  const pathName = location.pathname
  const navigator = useNavigate()

  useEffect(() => {
    if (pathName === '/semanticCache') {
      const queryParams = new URLSearchParams(location.search).toString()
      navigator(`/semanticCache/list${queryParams ? `?${queryParams}` : ''}`)
    }
  }, [pathName])
  return <Outlet></Outlet>
}
