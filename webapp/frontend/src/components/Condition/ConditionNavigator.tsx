import { IoIosArrowBack, IoIosArrowForward } from 'react-icons/io'
import { Condition, DEFAULT_CONDITION_LIMIT } from '/@/lib/apis'
import IconButton from '/@/components/UI/IconButton'

interface Props {
  conditions: Condition[]
  page: number
  next: () => Promise<void>
  prev: () => void
}

const ConditionNavigator = ({ conditions, next, prev, page }: Props) => {
  const isNextExist = conditions.length === DEFAULT_CONDITION_LIMIT
  const isPrevExist = page > 1

  return (
    <div className="center flex gap-8">
      <IconButton disabled={!isPrevExist} onClick={prev}>
        <IoIosArrowBack size={24} />
      </IconButton>
      <div className="align-middle text-xl">{page}</div>
      <IconButton disabled={!isNextExist} onClick={next}>
        <IoIosArrowForward size={24} />
      </IconButton>
    </div>
  )
}

export default ConditionNavigator
