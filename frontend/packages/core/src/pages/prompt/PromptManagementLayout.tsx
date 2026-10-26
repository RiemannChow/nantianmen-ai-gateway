import { useEffect } from 'react'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'

export default function PromptManagementLayout() {
  const location = useLocation()
  const pathName = location.pathname
  const navigator = useNavigate()

  useEffect(() => {
    if (pathName === '/prompt') {
      const queryParams = new URLSearchParams(location.search).toString()
      navigator(`/prompt/list${queryParams ? `?${queryParams}` : ''}`)
    }
  }, [pathName])
  return <Outlet></Outlet>
}
