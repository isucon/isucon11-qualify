import HomeCondition from '../components/Home/Condition'
import Isus from '../components/Home/Isus'
import Card from '../components/UI/Card'

const Home = () => {
  return (
    <div className="flex flex-col gap-10 items-center p-10">
      <Card>
        <Isus />
      </Card>
      <Card>
        <HomeCondition />
      </Card>
    </div>
  )
}

export default Home
