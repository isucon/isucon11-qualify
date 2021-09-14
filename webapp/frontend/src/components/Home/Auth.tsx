import Button from '/@/components/UI/Button'
import Card from '/@/components/UI/Card'

const Auth = () => {
  const click = async () => {
    location.href = `http://localhost:5000?callback=${location.protocol}//${location.host}`
  }

  return (
    <Card>
      <div className="flex flex-col items-center w-full">
        <img
          src="/assets/logo_orange.svg"
          alt="isucondition"
          className="my-2 max-w-sm"
        />
        <div className="mt-4 text-lg">ISUとつくる新しい明日</div>
        <div className="mt-16 w-full border-b" />
        <Button
          label="JIAのアカウントでログイン"
          customClass="mt-10 px-5 py-2 h-12 text-white font-bold bg-button rounded-3xl"
          onClick={click}
        />
      </div>
    </Card>
  )
}

export default Auth
