import { FaWeightHanging } from 'react-icons/fa'
import { MdBrokenImage } from 'react-icons/md'
import { GiSplash } from 'react-icons/gi'

interface Props {
  name: string
  status: boolean
}

const ConditionIcon = ({ name, status }: Props) => {
  const iconSize = 20
  const icon = (() => {
    switch (name) {
      case 'is_dirty':
        return <GiSplash size={iconSize} />
      case 'is_overweight':
        return <FaWeightHanging size={iconSize} />
      case 'is_broken':
        return <MdBrokenImage size={iconSize} />
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
