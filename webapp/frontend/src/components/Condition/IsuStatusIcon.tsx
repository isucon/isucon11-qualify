import { AiOutlineInfoCircle } from 'react-icons/ai'
import {
  BsFillExclamationTriangleFill,
  BsFillXOctagonFill
} from 'react-icons/bs'

interface Props {
  condition_level: 'info' | 'warning' | 'critical'
  size: number
}

const ConditionIcon = ({ condition_level, size }: Props) => {
  const icon = () => {
    switch (condition_level) {
      case 'info':
        return <AiOutlineInfoCircle size={size} />
      case 'warning':
        return <BsFillExclamationTriangleFill size={size} />
      case 'critical':
        return <BsFillXOctagonFill size={size} />
    }
  }
  const color = () => {
    switch (condition_level) {
      case 'info':
        return 'text-green-500'
      case 'warning':
        return 'text-yellow-500'
      case 'critical':
        return 'text-red-500'
    }
  }

  // TODO: tooltip
  return (
    <div className={'flex items-center bg-white z-1 pt-2 pb-2 ' + color()}>
      {icon()}
    </div>
  )
}

export default ConditionIcon
