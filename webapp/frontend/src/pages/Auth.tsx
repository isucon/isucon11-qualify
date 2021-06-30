import { useHistory } from 'react-router-dom'
import Card from '../components/UI/Card'
import { useDispatchContext } from '../context/state'
import apis, { debugGetJWT } from '../lib/apis'
import logo from '/@/assets/logo.png'

const Auth = () => {
  const dispatch = useDispatchContext()
  const history = useHistory()

  // TODO: 外部APIのログインページにリダイレクト
  const click = async () => {
    const jwt = await debugGetJWT()
    await apis.postAuth(jwt)

    const me = await apis.getUserMe()
    dispatch({ type: 'login', user: me })
    history.push('/')
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
