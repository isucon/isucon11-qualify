import Modal from 'react-modal'
import OptionKeyItem from './OptionKeyItem'
import { ALLOW_KEYS } from './use/parseQuery'

interface Props {
  isOpen: boolean
  rect: { x: number; y: number; width: number }
  toggle: () => void
  insertOption: (key: string) => void
}

const HelperModal = ({ isOpen, toggle, rect, insertOption }: Props) => {
  return (
    <Modal
      isOpen={isOpen}
      onRequestClose={toggle}
      style={{ content: { left: rect.x, top: rect.y, width: rect.width } }}
      className="bg-gray-50 absolute border border-outline rounded"
      overlayClassName="fixed inset-0"
      shouldCloseOnOverlayClick={true}
    >
      <div className="flex flex-col">
        {ALLOW_KEYS.map(v => (
          <OptionKeyItem
            key={v.key}
            keyName={v.key}
            description={v.description}
            onClick={() => insertOption(v.key)}
          />
        ))}
      </div>
    </Modal>
  )
}

export default HelperModal
