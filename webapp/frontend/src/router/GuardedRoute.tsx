import { Route, Redirect } from 'react-router-dom'
import { useStateContext } from '../context/state'

const GuardedRoute = <T extends { path: string; children?: JSX.Element }>(
  props: T
) => {
  const state = useStateContext()
  if (!state.me) {
    return <Redirect to="/login" />
  }

  return <Route {...props} />
}

export default GuardedRoute
