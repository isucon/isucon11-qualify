import Auth from '/@/components/Home/Auth'
import IsuList from '/@/components/Home/IsuList'
import TrendList from '/@/components/Home/TrendList'
import Card from '/@/components/UI/Card'
import { useStateContext } from '/@/context/state'

const Home = () => {
  const state = useStateContext()

  if (state.me) {
    return <LoginPage />
  } else {
    return <LandingPage />
  }
}

const LandingPage = () => {
  return (
    <div className="flex flex-col gap-10 items-center p-10 min-h-full">
      <Auth />
      <Card>
        <TrendList />
      </Card>
    </div>
  )
}

const LoginPage = () => {
  return (
    <div className="flex items-center justify-center p-10 min-w-full h-full">
      <Card>
        <IsuList />
      </Card>
    </div>
  )
}

export default Home
