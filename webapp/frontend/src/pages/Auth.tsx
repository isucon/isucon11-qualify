import { useDispatchContext } from '../context/state'
import apis, { debugGetJWT } from '../lib/apis'

const Auth = () => {
  const dispatch = useDispatchContext()

  // TODO: 外部APIのログインページにリダイレクト
  const click = async () => {
    const jwt = await debugGetJWT()
    await apis.postAuth(jwt)

    const me = await apis.getUserMe()
    dispatch({ type: 'login', user: me })
  }
  return (
    <div>
      <button onClick={click}>ISU協会でログイン</button>
    </div>
  )
}

export default Auth
