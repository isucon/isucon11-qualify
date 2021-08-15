import { IoIosArrowBack, IoIosArrowForward } from 'react-icons/io'
import { Condition, DEFAULT_CONDITION_LIMIT } from '/@/lib/apis'
import Button from '/@/components/UI/Button'

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
    <div className="flex gap-4">
      <Button label="Prev" disabled={!isPrevExist} onClick={prev}>
        <IoIosArrowBack size={24} />
      </Button>
      <div className="align-middle text-xl">{page}</div>
      <Button label="Next" disabled={!isNextExist} onClick={next}>
        <IoIosArrowForward size={24} />
      </Button>
    </div>
  )
}

export default ConditionNavigator
