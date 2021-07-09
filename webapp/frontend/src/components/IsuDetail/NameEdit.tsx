import { useState } from 'react'
import apis, { Isu } from '../../lib/apis'
import { HiOutlinePencil, HiOutlineCheck } from 'react-icons/hi'
import { IoMdClose } from 'react-icons/io'
import AutosizeInput from 'react-input-autosize'

interface Props {
  isu: Isu
  setIsu: React.Dispatch<React.SetStateAction<Isu | null>>
}
const NameEdit = ({ isu, setIsu }: Props) => {
  const [isEdit, setIsEdit] = useState(false)
  const [name, setName] = useState(isu.name)

  const startEdit = () => setIsEdit(true)
  const finishEdit = () => {
    setName(isu.name)
    setIsEdit(false)
  }
  const confirmEdit = async () => {
    await apis.putIsu(isu.jia_isu_uuid, { name })
    setIsEdit(false)
    setIsu({ ...isu, name })
  }
  return (
    <div className="flex">
      <AutosizeInput
        inputClassName="text-xl font-bold flex-1 mr-4"
        value={name}
        readOnly={!isEdit}
        onChange={e => setName(e.target.value)}
      />
      {isEdit ? (
        <FinishEditButtons confirmEdit={confirmEdit} finishEdit={finishEdit} />
      ) : (
        <StartEditButton startEdit={startEdit} />
      )}
    </div>
  )
}

const FinishEditButtons = ({
  confirmEdit,
  finishEdit
}: {
  confirmEdit: () => Promise<void>
  finishEdit: () => void
}) => {
  return (
    <div className="flex gap-2 items-center">
      <button
        onClick={confirmEdit}
        className="flex items-center focus:outline-none"
      >
        <HiOutlineCheck size="20" />
      </button>
      <button
        onClick={finishEdit}
        className="flex items-center text-error focus:outline-none"
      >
        <IoMdClose size="20" />
      </button>
    </div>
  )
}

const StartEditButton = ({ startEdit }: { startEdit: () => void }) => {
  return (
    <button
      onClick={startEdit}
      className="flex items-center focus:outline-none"
    >
      <HiOutlinePencil size="20" />
    </button>
  )
}

export default NameEdit
