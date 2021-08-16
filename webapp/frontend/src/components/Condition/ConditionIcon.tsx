import { FaWeightHanging } from 'react-icons/fa'
import { BsWrench } from 'react-icons/bs'
import { AiOutlineClear } from 'react-icons/ai'
interface Props {
  name: string
  status: boolean
}

const ConditionIcon = ({ name, status }: Props) => {
  const icon = () => {
    switch (name) {
      case 'is_dirty':
        return <AiOutlineClear size={24} />
      case 'is_overweight':
        return <FaWeightHanging size={18} />
      case 'is_broken':
        return <BsWrench size={20} />
    }
  }

  return (
    <div>
      <div
        data-tip={name}
        data-place="top"
        className={
          'flex items-center text-primary ' + (status || ' opacity-30')
        }
      >
        {icon()}
      </div>
    </div>
  )
}

export default ConditionIcon
