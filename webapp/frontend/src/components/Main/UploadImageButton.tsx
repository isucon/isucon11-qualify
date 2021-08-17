import { useEffect } from 'react'
import Button from '/@/components/UI/Button'

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
    <Button
      customClass="px-3 py-1 h-8 leading-4 border rounded"
      label="画像をアップロード"
      onClick={startSelect}
    />
  )
}

export default IconInput
