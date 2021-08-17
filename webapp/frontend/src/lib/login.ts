import { CancelToken } from 'axios'
import { useCallback, useState } from 'react'
import { useDispatchContext, useStateContext } from '/@/context/state'
import apis, { User } from './apis'
import toast from 'react-hot-toast'

const useLogin = () => {
  const state = useStateContext()
  const dispatch = useDispatchContext()
  const [isTryLogin, setTryLogin] = useState(true)

  const setMe = useCallback(
    (me: User) => {
      dispatch({ type: 'login', user: me })
      setTryLogin(false)
    },
    [dispatch]
  )

  const login = useCallback(
    async (cancelToken?: CancelToken) => {
      if (state.me) {
        setTryLogin(false)
        return
      }
      try {
        setTryLogin(true)
        try {
          const me = await apis.getUserMe({ cancelToken })
          setMe(me)
        } catch (e) {
          // cookieがついてないとき
          const url = new URL(location.href)
          const jwt = url.searchParams.get('jwt')
          if (!jwt) {
            setTryLogin(false)
            toast.error(e.response.data)
            return
          }
          url.searchParams.delete('jwt')
          history.replaceState(null, '', url.href)
          await apis.postAuth(jwt, { cancelToken })
          const me = await apis.getUserMe({ cancelToken })
          setMe(me)
        }
      } catch (e) {
        toast.error(e.response.data)
        setTryLogin(false)
      }
    },
    [state.me, setMe]
  )

  return { isTryLogin, login, state }
}

export default useLogin
