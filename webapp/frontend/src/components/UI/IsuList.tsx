import { GetIsuListResponse } from '/@/lib/apis'
import { getConditionTime } from '/@/lib/date'
import Isu from './Isu'
import Tip from './Tip'

interface Props {
  isus: GetIsuListResponse[]
}

const IsuList = ({ isus }: Props) => {
  return (
    <div className="grid gap-8 grid-cols-isus w-full">
      {isus.map(isu => (
        <div key={isu.jia_isu_uuid} className="flex flex-col items-center">
          <Isu isu={isu} />
          {isu.latest_isu_condition ? (
            <div>{getConditionTime(isu.latest_isu_condition.date)}</div>
          ) : null}
          {isu.latest_isu_condition ? (
            <Tip variant={isu.latest_isu_condition.condition_level} />
          ) : null}
        </div>
      ))}
    </div>
  )
}

export default IsuList
