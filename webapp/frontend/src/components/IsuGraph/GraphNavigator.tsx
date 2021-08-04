import { IoIosArrowBack, IoIosArrowForward } from 'react-icons/io'
import IconButton from '../UI/IconButton'
import DateInput from './DateInput'

interface Props {
  next: () => Promise<void>
  prev: () => void
  day: string
  fetchGraphs: (payload: { day: string }) => Promise<void>
}

const ConditionNavigator = ({ next, prev, day, fetchGraphs }: Props) => {
  return (
    <div className="flex gap-8">
      <IconButton onClick={prev}>
        <IoIosArrowBack size={24} />
      </IconButton>
      <DateInput day={day} fetchGraphs={fetchGraphs} />
      <IconButton onClick={next}>
        <IoIosArrowForward size={24} />
      </IconButton>
    </div>
  )
}

export default ConditionNavigator
