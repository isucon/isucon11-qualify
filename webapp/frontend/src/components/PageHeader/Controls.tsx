import { Link } from 'react-router-dom'
import { useStateContext } from '../../context/state'

const Controls = () => {
  const me = useStateContext().me
  if (!me) {
    return <></>
  }
  return (
    <div className="flex ml-auto">
      <Link to="/isu/condition">
        <div>ISUの状態</div>
      </Link>
      <Link to="/isu/search">
        <div>ISUの検索</div>
      </Link>
      <div>{me.jiaUserID}</div>
    </div>
  )
}

export default Controls
