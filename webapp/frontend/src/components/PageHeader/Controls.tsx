import { Link } from 'react-router-dom'
import { useStateContext } from '../../context/state'
import { MdAccountCircle } from 'react-icons/md'
import { useState } from 'react'

import ControlItem from './ControlItem'
import UserControlModal from './UserControlModal'

const Controls = () => {
  const [isOpenModal, setIsOpenModal] = useState(false)
  const toggleModal = () => {
    setIsOpenModal(!isOpenModal)
  }
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
        <div className="flex items-center cursor-pointer" onClick={toggleModal}>
          <MdAccountCircle />
          <div>{me.jiaUserID}</div>
        </div>
      </ControlItem>
      <UserControlModal isOpen={isOpenModal} toggle={toggleModal} />
    </div>
  )
}

export default Controls
