import { useCallback, useState } from 'react'
import { useDispatchContext, useStateContext } from '../context/state'
import apis from './apis'

const useLogin = () => {
  const state = useStateContext()
  const dispatch = useDispatchContext()
  const [isTryLogin, setTryLogin] = useState(true)

  const login = useCallback(async () => {
    setTryLogin(true)
    try {
      if (!state.me) {
        try {
          const me = await apis.getUserMe()
          dispatch({ type: 'login', user: me })
        } catch {
          // cookieがついてないとき
          const url = new URL(location.href)
          const jwt = url.searchParams.get('jwt')
          if (jwt) {
            await apis.postAuth(jwt)
            const me = await apis.getUserMe()
            dispatch({ type: 'login', user: me })
          }
        }
      }
    } finally {
      setTryLogin(false)
    }
  }, [state.me, dispatch])

  return { isTryLogin, login, state }
}

export default useLogin
