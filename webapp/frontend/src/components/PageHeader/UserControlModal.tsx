import Modal from 'react-modal'
import { IoMdLogOut } from 'react-icons/io'
import { Link } from 'react-router-dom'
import { useDispatchContext } from '/@/context/state'
import apis from '/@/lib/apis'

interface Props {
  isOpen: boolean
  toggle: () => void
}

const UserControlModal = (props: Props) => {
  const dispatch = useDispatchContext()
  const logout = async () => {
    await apis.postSignout()
    dispatch({ type: 'logout' })
  }

  return (
    <Modal
      isOpen={props.isOpen}
      onRequestClose={props.toggle}
      className="top-18 bg-gray-50 absolute right-0 flex justify-items-center px-6 py-3 border rounded"
      overlayClassName="fixed inset-0"
      shouldCloseOnOverlayClick={true}
    >
      <Link to="/" onClick={logout} className="text-primary flex items-center">
        <IoMdLogOut className="mr-1" size={16} />
        <div>logout</div>
      </Link>
    </Modal>
  )
}

export default UserControlModal
