import { IoIosArrowBack, IoIosArrowForward } from 'react-icons/io'
import { Condition, DEFAULT_CONDITION_LIMIT } from '../../lib/apis'
import IconButton from '../UI/IconButton'
import ConditionDetail from './ConditionDetail'

interface Props {
  conditions: Condition[]
  page: number
  next: () => Promise<void>
  prev: () => void
}

const Conditions = ({ conditions, next, prev, page }: Props) => {
  const isNextExist = conditions.length === DEFAULT_CONDITION_LIMIT
  const isPrevExist = page > 1

  return (
    <div className="flex flex-col gap-4 items-center">
      <div className="w-full border border-b-0 border-outline">
        {conditions.map((condition, i) => (
          <div className="border-b border-outline" key={i}>
            <ConditionDetail condition={condition} />
          </div>
        ))}
      </div>
      <div className="center flex gap-8">
        <IconButton disabled={!isPrevExist} onClick={prev}>
          <IoIosArrowBack size={24} />
        </IconButton>
        <div className="align-middle text-xl">{page}</div>
        <IconButton disabled={!isNextExist} onClick={next}>
          <IoIosArrowForward size={24} />
        </IconButton>
      </div>
    </div>
  )
}

export default Conditions
