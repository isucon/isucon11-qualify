import { Route, Redirect } from 'react-router-dom'

const GuardedRoute = <T extends { path: string; children?: JSX.Element }>(
  props: T
) => {
  // TODO: ここで GET /user/me をたたく or storeに問い合わせ
  const me = null
  if (!me) {
    return <Redirect to="/login" />
  }

  return <Route {...props} />
}

export default GuardedRoute
