import IsuConditionCardContent from '../components/Condition/IsuConditionCardContent'
import Card from '../components/UI/Card'
import NowLoading from '../components/UI/NowLoading'
import { Isu } from '../lib/apis'

interface Props {
  isu: Isu
}

const IsuCondition = ({ isu }: Props) => {
  if (!isu) {
    return <NowLoading />
  }
  return (
    <div className="flex flex-col gap-10 items-center">
      <Card>
        <IsuConditionCardContent isu={isu} />
      </Card>
    </div>
  )
}

export default IsuCondition
