import { useEffect } from 'react'
import apis, { Isu } from '../../lib/apis'

interface Props {
  isu: Isu
  reloadIcon?: () => void
}

const useImageSelect = (onSelect: (file: File) => void) => {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/png'

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

const IconInput = ({ isu, reloadIcon }: Props) => {
  const putIsuIcon = async (file: File) => {
    await apis.putIsuIcon(isu.jia_isu_uuid, file)
    if (reloadIcon) {
      reloadIcon()
    }
  }
  const { startSelect, destroy } = useImageSelect(putIsuIcon)
  useEffect(() => destroy)

  return <button onClick={startSelect}>画像をアップロード</button>
}

export default IconInput
