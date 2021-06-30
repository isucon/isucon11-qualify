import { IoIosArrowBack, IoIosArrowForward } from 'react-icons/io'
import IconButton from '../UI/IconButton'

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

export default PagingNavigator
