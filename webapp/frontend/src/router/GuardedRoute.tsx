import { useEffect } from 'react'
import { Route, Redirect } from 'react-router-dom'
import NowLoading from '../components/UI/NowLoading'
import useLogin from '../lib/login'

const GuardedRoute = <T extends { path: string; children?: JSX.Element }>(
  props: T
) => {
  const { isTryLogin, login, state } = useLogin()
  useEffect(() => {
    login()
  }, [login])

  if (isTryLogin) {
    return <NowLoading />
  }

  if (!state.me) {
    return <Redirect to={`/login`} />
  }

  return <Route {...props} />
}

export default GuardedRoute
