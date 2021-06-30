import Card from '../components/UI/Card'
import NowLoading from '../components/UI/NowLoading'
import { Isu } from '../lib/apis'

interface Props {
  isu: Isu
  setIsu: React.Dispatch<React.SetStateAction<Isu | null>>
}

const IsuCondition = ({ isu }: Props) => {
  if (!isu) {
    return <NowLoading />
  }
  return (
    <div className="flex flex-col gap-10 items-center">
      <Card>
        <div>nanka</div>
      </Card>
    </div>
  )
}

export default IsuCondition
