import Card from '/@/components/UI/Card'
import { Isu } from '/@/lib/apis'
import IsuGraphCardContent from '/@/components/IsuGraph/IsuGraphCardContent'

interface Props {
  isu: Isu
}

const IsuGraph = ({ isu }: Props) => {
  return (
    <div className="flex flex-col gap-10 items-center">
      <Card>
        <IsuGraphCardContent isu={isu} />
      </Card>
    </div>
  )
}

export default IsuGraph
