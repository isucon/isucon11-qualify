import Modal from 'react-modal'
import { IoMdLogOut } from 'react-icons/io'
import { Link } from 'react-router-dom'
import { useDispatchContext } from '../../context/state'

interface Props {
  isOpen: boolean
  toggle: () => void
}

const UserControlModal = (props: Props) => {
  const dispatch = useDispatchContext()
  const logout = () => {
    dispatch({ type: 'logout' })
  }

  return (
    <Modal
      isOpen={props.isOpen}
      onRequestClose={props.toggle}
      className="right-8 top-8 border-dark-200 bg-gray-50 absolute p-4 w-40 border rounded"
      overlayClassName="fixed inset-0"
      shouldCloseOnOverlayClick={true}
    >
      <Link
        to="/"
        onClick={logout}
        className="flex items-center text-primary-800"
      >
        <IoMdLogOut className="mr-3" size={20} />
        <div>logout</div>
      </Link>
    </Modal>
  )
}

export default UserControlModal
