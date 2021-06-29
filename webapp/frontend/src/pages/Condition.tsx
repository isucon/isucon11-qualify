import { useEffect } from 'react'
import { useState } from 'react'
import Conditions from '../components/Condition/Conditions'
import SearchInputs from '../components/Condition/SearchInputs'
import Card from '../components/UI/Card'
import apis, { Condition } from '../lib/apis'

const ConditionComponent = () => {
  const [conditions, setConditions] = useState<Condition[]>([])
  useEffect(() => {
    const fetchCondtions = async () => {
      setConditions(
        await apis.getConditions({
          cursor_end_time: new Date(),
          // 初回fetch時は'z'をセットすることで全件表示させてる
          cursor_jia_isu_uuid: 'z',
          condition_level: 'critical,warning,info'
        })
      )
    }
    fetchCondtions()
  }, [setConditions])

  return (
    <div className="p-10">
      <Card>
        <div className="flex flex-col gap-2">
          <h2 className="text-xl font-bold">Condition</h2>
          <SearchInputs />
          <Conditions conditions={conditions} />
        </div>
      </Card>
    </div>
  )
}

export default ConditionComponent
