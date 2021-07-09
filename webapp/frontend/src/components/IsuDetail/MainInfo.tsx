import { useState } from 'react'
import { useHistory } from 'react-router-dom'
import apis, { Isu } from '../../lib/apis'
import IconInput from '../UI/IconInput'
import IsuIcon from './IsuIcon'
import NameEdit from './NameEdit'

interface Props {
  isu: Isu
  setIsu: React.Dispatch<React.SetStateAction<Isu | null>>
}

const MainInfo = ({ isu, setIsu }: Props) => {
  const history = useHistory()

  const deleteIsu = async () => {
    if (isu && confirm(`本当に${isu.name}を削除しますか？`)) {
      await apis.deleteIsu(isu.jia_isu_uuid)
      history.push('/')
    }
  }

  const [iconKey, setIconKey] = useState(0)
  const putIsuIcon = async (file: File) => {
    await apis.putIsuIcon(isu.jia_isu_uuid, file)
    setIconKey(performance.now())
  }
  return (
    <div className="flex flex-wrap gap-16 justify-center">
      <IsuIcon isu={isu} reloadKey={iconKey} />
      <div className="flex flex-col min-h-full">
        <NameEdit isu={isu} setIsu={setIsu} />
        <div className="flex flex-1 flex-col mt-4 pl-8">
          <div className="flex-1">{isu.character}</div>
          <div className="flex flex-no-wrap gap-4 justify-self-end mt-12">
            <IconInput putIsuIcon={putIsuIcon} />
            <button
              className="px-3 py-1 text-error border border-error rounded"
              onClick={deleteIsu}
            >
              削除
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

export default MainInfo
