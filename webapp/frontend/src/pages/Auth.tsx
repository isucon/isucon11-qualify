import { useDispatchContext } from '../context/state'
import apis from '../lib/apis'
import { useEffect } from 'react'

const Auth = () => {
  const dispatch = useDispatchContext()
  useEffect(() => {
    const login = async () => {
      const url = new URL(location.href)
      const jwt = url.searchParams.get('jwt')
      if (jwt) {
        await apis.postAuth(jwt)

        const me = await apis.getUserMe()
        dispatch({ type: 'login', user: me })
      }
    }
    login()
  })
  const click = async () => {
    // TODO: 本番どうするか考える
    location.href = `http://localhost:5000`
  }

  return (
    <div>
      <button onClick={click}>ISU協会でログイン</button>
    </div>
  )
}

export default Auth
