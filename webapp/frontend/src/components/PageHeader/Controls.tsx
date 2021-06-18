import { Link } from 'react-router-dom'
import { useStateContext } from '../../context/state'
import { MdAccountCircle } from 'react-icons/md'
import ControlItem from './ControlItem'

const Controls = () => {
  const me = useStateContext().me
  if (!me) {
    return <></>
  }
  return (
    <div className="flex ml-auto items-center">
      <ControlItem>
        <Link to="/isu/condition">
          <div>ISUの状態</div>
        </Link>
      </ControlItem>
      <ControlItem>
        <Link to="/isu/search">
          <div>ISUの検索</div>
        </Link>
      </ControlItem>
      <ControlItem>
        <div className="flex items-center cursor-pointer">
          <MdAccountCircle />
          <div>{me.jiaUserID}</div>
        </div>
      </ControlItem>
    </div>
  )
}

export default Controls
