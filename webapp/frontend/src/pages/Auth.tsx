import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useDispatchContext } from '../context/state'
import apis, { debugGetJWT, Isu } from '../lib/apis'

const Auth = () => {
  const dispatch = useDispatchContext()
  const [isus, setIsus] = useState<Isu[]>([])

  // TODO: 外部APIのログインページにリダイレクト
  const click = async () => {
    const jwt = await debugGetJWT()
    await apis.postAuth(jwt)

    const me = await apis.getUserMe()
    setIsus(await apis.getIsus())
    dispatch({ type: 'login', user: me })
  }
  return (
    <div>
      <button onClick={click}>ISU協会でログイン</button>
      {isus.map(isu => {
        return (
          <Link
            to={`/isu/${isu.jia_isu_uuid}`}
            key={isu.jia_isu_uuid}
            className="mr-2"
          >
            {isu.name}
          </Link>
        )
      })}
    </div>
  )
}

export default Auth
