import apis from '../../lib/apis'
import Conditions from '../Condition/Conditions'
import usePaging from '../Condition/use/paging'

const HomeCondition = () => {
  const { conditions } = usePaging(apis.getConditions)

  return (
    <div className="flex flex-col gap-2">
      <h2 className="text-xl font-bold">Condition</h2>
      <div className="flex flex-col gap-4 items-center">
        <Conditions conditions={conditions} />
      </div>
    </div>
  )
}

export default HomeCondition
