import { useStateContext } from '../../context/state'
import { MdAccountCircle } from 'react-icons/md'
import { useState } from 'react'

import ControlItem from './ControlItem'
import UserControlModal from './UserControlModal'
import ControlLinkItem from './ControlLinkItem'
import { IoIosNotifications, IoMdSearch } from 'react-icons/io'

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
    <div className="flex items-center ml-auto">
      <ControlLinkItem
        to="/isu/condition"
        label="ISUの状態"
        icon={<IoIosNotifications />}
      />
      <ControlLinkItem
        to="/isu/search"
        label="ISUの検索"
        icon={<IoMdSearch />}
      />
      <ControlItem>
        <div className="flex items-center cursor-pointer" onClick={toggleModal}>
          <MdAccountCircle />
          <div className="ml-1">{me.jia_user_id}</div>
        </div>
      </ControlItem>
      <UserControlModal isOpen={isOpenModal} toggle={toggleModal} />
    </div>
  )
}

export default Controls
