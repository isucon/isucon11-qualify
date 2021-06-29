import { useDispatchContext } from '../context/state'
import apis, { debugGetJWT } from '../lib/apis'
import { Link } from 'react-router-dom'

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
      <Link to="/isu/0694e4d7-dfce-4aec-b7ca-887ac42cfb8f">hoge</Link>
    </div>
  )
}

export default Auth
