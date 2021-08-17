import { IoIosArrowBack, IoIosArrowForward } from 'react-icons/io'
import Button from '/@/components/UI/Button'
import DateInput from './DateInput'

interface Props {
  next: () => Promise<void>
  prev: () => void
  specify: (day: string) => Promise<void>
  day: string
}

const ConditionNavigator = ({ next, prev, specify, day }: Props) => {
  return (
    <div className="flex gap-8">
      <Button label="Prev" onClick={prev}>
        <IoIosArrowBack size={24} />
      </Button>
      <DateInput day={day} setDay={specify} />
      <Button label="Next" onClick={next}>
        <IoIosArrowForward size={24} />
      </Button>
    </div>
  )
}

export default ConditionNavigator
