import HomeCondition from '../components/Home/Condition'
import Card from '../components/UI/Card'

const Home = () => {
  return (
    <div className="flex flex-col gap-10 items-center p-10">
      <Card>
        <div>aikonntoka</div>
      </Card>
      <Card>
        <HomeCondition />
      </Card>
    </div>
  )
}

export default Home
