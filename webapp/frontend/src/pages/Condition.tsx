import SearchInputs from '../components/Condition/SearchInputs'
import Card from '../components/UI/Card'

const Condition = () => {
  return (
    <div className="p-10">
      <Card>
        <div className="flex flex-col gap-2">
          <h2 className="text-xl font-bold">Condition</h2>
          <SearchInputs />
        </div>
      </Card>
    </div>
  )
}

export default Condition
