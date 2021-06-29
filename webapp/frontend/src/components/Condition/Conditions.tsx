import { IoIosArrowBack, IoIosArrowForward } from 'react-icons/io'
import { Condition } from '../../lib/apis'
import IconButton from '../UI/IconButton'
import ConditionDetail from './ConditionDetail'

interface Props {
  conditions: Condition[]
}

const DEFAULT_CONDITION_LIMIT = 20

const Conditions = ({ conditions }: Props) => {
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
        <IconButton>
          <IoIosArrowBack size={24} />
        </IconButton>
        <div className="align-middle text-xl">1</div>
        <IconButton>
          <IoIosArrowForward size={24} />
        </IconButton>
      </div>
    </div>
  )
}

export default Conditions
