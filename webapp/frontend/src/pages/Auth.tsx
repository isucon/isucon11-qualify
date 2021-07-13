import Card from '../components/UI/Card'
import logo_orange from '/@/assets/logo_orange.svg'

const Auth = () => {
  const click = async () => {
    // TODO: 本番どうするか考える
    location.href = `http://localhost:5000`
  }

  return (
    <div className="flex justify-center p-10">
      <div className="flex justify-center w-full max-w-lg">
        <Card>
          <div className="flex flex-col items-center w-full">
            <img src={logo_orange} alt="isucondition" className="my-2" />
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
