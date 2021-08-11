import Auth from '../components/Home/Auth'
import Isus from '../components/Home/Isus'
import TrendList from '../components/Home/TrendList'
import Card from '../components/UI/Card'
import { useStateContext } from '../context/state'

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
    <div className="flex flex-col gap-10 items-center p-10">
      <Auth />
      <Card>
        <TrendList />
      </Card>
    </div>
  )
}

const LoginPage = () => {
  return (
    <div className="flex flex-col gap-10 items-center p-10">
      <Card>
        <Isus />
      </Card>
    </div>
  )
}

export default Home
