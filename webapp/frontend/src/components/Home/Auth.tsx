import Card from '../UI/Card'
import logo_orange from '/@/assets/logo_orange.svg'

const Auth = () => {
  const click = async () => {
    location.href = `${location.protocol}//${document.domain}:5000?callback=${location.protocol}//${location.host}`
  }

  return (
    <Card>
      <div className="flex flex-col items-center w-full">
        <img src={logo_orange} alt="isucondition" className="my-2 max-w-sm" />
        <div className="mt-4 text-lg">ISUとつくる新しい明日</div>
        <div className="mt-16 w-full border-b border-outline" />
        <button
          className="mt-10 px-5 py-2 h-12 text-white-primary font-bold bg-button rounded-3xl"
          onClick={click}
        >
          JIAのアカウントでログイン
        </button>
      </div>
    </Card>
  )
}

export default Auth
