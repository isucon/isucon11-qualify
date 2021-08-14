import { Isu } from '../../lib/apis'
import IsuIcon from './IsuIcon'

interface Props {
  isu: Isu
}

const MainInfo = ({ isu }: Props) => {
  return (
    <div className="flex flex-wrap gap-16 justify-center">
      <IsuIcon isu={isu} />
      <div className="flex flex-col min-h-full">
        <div className="text-xl font-bold">{isu.name}</div>
        <div className="flex flex-1 mt-4 pl-8">
          <div className="mr-4">せいかく</div>
          <div className="flex-1">{isu.character}</div>
        </div>
      </div>
    </div>
  )
}

export default MainInfo
