import { Link } from 'react-router-dom'
import ControlItem from './ControlItem'

interface Props {
  to: string
  label: string
  icon: JSX.Element
}

const ControlLinkItem = (props: Props) => {
  return (
    <ControlItem>
      <Link to={props.to} className="flex items-center">
        {props.icon}
        <div className="ml-1">{props.label}</div>
      </Link>
    </ControlItem>
  )
}

export default ControlLinkItem
