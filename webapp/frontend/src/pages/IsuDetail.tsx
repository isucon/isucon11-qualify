import IsuImage from '/@/components/UI/IsuImage'
import { Isu } from '/@/lib/apis'

interface Props {
  isu: Isu
}

const IsuDetail = ({ isu }: Props) => {
  return (
    <div className="flex flex-wrap gap-16">
      <IsuImage isu={isu} customClass="h-64 w-64" />
      <div className="flex flex-col flex-grow">
        <div className="mb-4 text-xl font-bold">{isu.name}</div>
        <div className="grid-cols-[max-content,max-content,max-content] grid gap-2 ml-4">
          <div>性格:</div>
          <div>{isu.character}</div>
        </div>
      </div>
    </div>
  )
}

export default IsuDetail
