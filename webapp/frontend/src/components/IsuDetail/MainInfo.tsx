<<<<<<< HEAD
import { useHistory } from 'react-router-dom'
import apis, { Isu } from '../../lib/apis'
=======
import { Isu } from '/@/lib/apis'
>>>>>>> 08.15.1
import IsuIcon from './IsuIcon'

interface Props {
  isu: Isu
}

const MainInfo = ({ isu }: Props) => {
  const history = useHistory()

  const deleteIsu = async () => {
    if (isu && confirm(`本当に${isu.name}を削除しますか？`)) {
      await apis.deleteIsu(isu.jia_isu_uuid)
      history.push('/')
    }
  }

  return (
    <div className="flex flex-wrap gap-16 justify-center">
      <IsuIcon isu={isu} />
      <div className="flex flex-col min-h-full">
        <div className="text-xl font-bold">{isu.name}</div>
        <div className="flex flex-1 flex-col mt-4 pl-8">
          <div className="flex-1">{isu.character}</div>
          <div className="flex flex-no-wrap gap-4 justify-self-end mt-12">
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
