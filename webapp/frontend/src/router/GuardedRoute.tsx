import { useState } from 'react'
import { useEffect } from 'react'
import { Route, Redirect } from 'react-router-dom'
import NowLoading from '../components/UI/NowLoading'
import { useStateContext } from '../context/state'
import useLogin from '../lib/login'

const GuardedRoute = <T extends { path: string; children?: JSX.Element }>(
  props: T
) => {
  const state = useStateContext()
  const login = useLogin(state)
  const [isTryLogin, setIsTryLogin] = useState(true)
  useEffect(() => {
    login().then(() => {
      setIsTryLogin(false)
    })
  }, [login])

  if (isTryLogin) {
    return <NowLoading />
  }

  if (!state.me) {
    const url = new URL(location.href)
    return <Redirect to={`/login${url.search}`} />
  }

  return <Route {...props} />
}

export default GuardedRoute
