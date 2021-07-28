import { useEffect } from 'react'

interface Props {
  putIsuIcon: (file: File) => void
}

const useImageSelect = (onSelect: (file: File) => void) => {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/jpeg'

  const onChange = () => {
    if (input.files && input.files[0]) {
      onSelect(input.files[0])
    }
  }

  input.addEventListener('change', onChange)

  const startSelect = () => {
    input.click()
  }

  const destroy = () => {
    input.removeEventListener('change', onChange)
  }

  return { startSelect, destroy }
}

const IconInput = ({ putIsuIcon }: Props) => {
  const { startSelect, destroy } = useImageSelect(putIsuIcon)
  useEffect(() => destroy)

  return (
    <button
      className="px-3 py-1 border border-outline rounded"
      onClick={startSelect}
    >
      画像をアップロード
    </button>
  )
}

export default IconInput
