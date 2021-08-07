import { FaWeightHanging } from 'react-icons/fa'
import { MdBrokenImage } from 'react-icons/md'
import { GiSplash } from 'react-icons/gi'

interface Props {
  name: string
  status: boolean
}

const ConditionIcon = ({ name, status }: Props) => {
  const icon = (() => {
    switch (name) {
      case 'is_dirty':
        return <GiSplash />
      case 'is_overweight':
        return <FaWeightHanging />
      case 'is_broken':
        return <MdBrokenImage />
    }
  })()

  const color = (() => {
    switch (status) {
      case true:
        return 'text-primary'
      case false:
        return 'text-teritary'
    }
  })()

  const className = `${color}`
  return <div className={className}>{icon}</div>
}

export default ConditionIcon
