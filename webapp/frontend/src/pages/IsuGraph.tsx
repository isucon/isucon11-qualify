import Card from '../components/UI/Card'
import NowLoading from '../components/UI/NowLoading'
import { Isu } from '../lib/apis'
import IsuGraphCardContent from '../components/IsuGraph/IsuGraphCardContent'

interface Props {
  isu: Isu
}

const IsuGraph = ({ isu }: Props) => {
  if (!isu) {
    return <NowLoading />
  }
  return (
    <div className="flex flex-col gap-10 items-center">
      <Card>
        <IsuGraphCardContent isu={isu} />
      </Card>
    </div>
  )
}

export default IsuGraph
