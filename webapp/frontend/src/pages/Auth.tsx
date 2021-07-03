import { useHistory } from 'react-router-dom'
import Card from '../components/UI/Card'
import apis from '../lib/apis'
import { useEffect } from 'react'
import logo from '/@/assets/logo.png'
import { useDispatchContext } from '../context/state'

const Auth = () => {
  const dispatch = useDispatchContext()
  const history = useHistory()

  useEffect(() => {
    const login = async () => {
      try {
        const me = await apis.getUserMe()
        dispatch({ type: 'login', user: me })
        history.push('/')
      } catch {
        const url = new URL(location.href)
        const jwt = url.searchParams.get('jwt')
        if (jwt) {
          await apis.postAuth(jwt)
          const me = await apis.getUserMe()
          dispatch({ type: 'login', user: me })
        }
        history.push('/')
      }
    }
    login()
  }, [history, dispatch])
  const click = async () => {
    // TODO: 本番どうするか考える
    location.href = `http://localhost:5000`
  }

  return (
    <div className="flex justify-center p-10">
      <div className="flex justify-center w-full max-w-lg">
        <Card>
          <div className="flex flex-col items-center w-full">
            <img src={logo} alt="isucondition" />
            <div className="mt-4 text-lg">ISUとつくる新しい明日</div>
            <div className="mt-6 w-full border-b border-outline" />
            <button
              className="mt-10 px-5 py-2 h-12 text-white-primary font-bold bg-button rounded-3xl"
              onClick={click}
            >
              JIAのアカウントでログイン
            </button>
          </div>
        </Card>
      </div>
    </div>
  )
}

export default Auth
