import IsuImage from '/@/components/UI/IsuImage'
import { Isu } from '/@/lib/apis'

interface Props {
  isu: Isu
}

const IsuDetail = ({ isu }: Props) => {
  return (
    <div className="flex flex-wrap gap-16 justify-center">
      <IsuImage isu={isu} customClass="h-64 w-64" />
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

export default IsuDetail
