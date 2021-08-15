import { useStateContext } from '/@/context/state'
import { MdAccountCircle } from 'react-icons/md'
import { useState } from 'react'

import ControlItem from './ControlItem'
import UserControlModal from './UserControlModal'
import ControlLinkItem from './ControlLinkItem'
import { IoIosNotifications } from 'react-icons/io'
import { TiPlus } from 'react-icons/ti'

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
    <div className="w-80 flex items-center justify-between ml-auto">
      <ControlLinkItem to="/register" label="ISUの登録" icon={<TiPlus />} />
      <ControlLinkItem
        to="/condition"
        label="ISUの状態"
        icon={<IoIosNotifications />}
      />
      {/* TODO: ISU一覧へのリンク */}
      <div className="border-l-1 pl-4 border-white">
        <ControlItem>
          <div
            className="flex items-center cursor-pointer"
            onClick={toggleModal}
          >
            <MdAccountCircle />
            <div className="ml-1">{me.jia_user_id}</div>
          </div>
        </ControlItem>
      </div>
      <UserControlModal isOpen={isOpenModal} toggle={toggleModal} />
    </div>
  )
}

export default Controls
