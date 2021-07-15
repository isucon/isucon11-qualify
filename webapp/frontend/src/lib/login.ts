import { useCallback } from 'react'
import { useDispatchContext, State } from '../context/state'
import apis from './apis'

const useLogin = (state: State) => {
  const dispatch = useDispatchContext()

  const login = useCallback(async () => {
    if (!state.me) {
      try {
        const me = await apis.getUserMe()
        dispatch({ type: 'login', user: me })
      } catch {
        const url = new URL(location.href)
        const jwt = url.searchParams.get('jwt')
        if (jwt) {
          await apis.postAuth(jwt)
          const me = await apis.getUserMe()
          dispatch({ type: 'login', user: me })
        } else {
          throw 'has no jwt'
        }
      }
    }
  }, [state, dispatch])

  return login
}

export default useLogin
