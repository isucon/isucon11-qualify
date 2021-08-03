import axios from 'axios'
import { useEffect } from 'react'
import { Route, Redirect } from 'react-router-dom'
import NowLoading from '../components/UI/NowLoading'
import useLogin from '../lib/login'

const GuardedRoute = <T extends { path: string; children?: JSX.Element }>(
  props: T
) => {
  const { isTryLogin, login, state } = useLogin()
  useEffect(() => {
    const cancelToken = axios.CancelToken
    const source = cancelToken.source()
    login(source.token)
    return () => source.cancel()
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
