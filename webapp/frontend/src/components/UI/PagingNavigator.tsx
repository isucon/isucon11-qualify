import { IoIosArrowBack, IoIosArrowForward } from 'react-icons/io'
import Button from '/@/components/UI/Button'

interface Props {
  length: number
  maxLength: number
  page: number
  next: () => Promise<void> | void
  prev: () => Promise<void> | void
}

const PagingNavigator = ({ length, maxLength, next, prev, page }: Props) => {
  const isNextExist = length === maxLength
  const isPrevExist = page > 1

  return (
    <div className="center flex gap-8">
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

export default PagingNavigator
