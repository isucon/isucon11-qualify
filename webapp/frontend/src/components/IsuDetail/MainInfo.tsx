import { useState } from 'react'
import { useHistory } from 'react-router-dom'
import apis, { Isu } from '../../lib/apis'
import IconInput from '../UI/IconInput'
import IsuIcon from './IsuIcon'

const MainInfo = ({ isu }: { isu: Isu }) => {
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
    <div className="flex flex-wrap gap-16 justify-center px-16 py-12">
      <IsuIcon isu={isu} reloadKey={iconKey} />
      <div className="flex flex-col min-h-full">
        <h2 className="mb-3 text-xl font-bold">{isu.name}</h2>
        <div className="flex flex-1 flex-col pl-6">
          <div className="flex-1">{isu.character}</div>
          <div className="flex flex-no-wrap gap-4 justify-self-end mt-12">
            <IconInput putIsuIcon={putIsuIcon} />
            <button
              className="px-3 py-1 text-white-primary bg-button rounded"
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
