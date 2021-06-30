import { useState, useEffect } from 'react'
import apis, { Condition } from '../../../lib/apis'

const useConditions = () => {
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

  return { conditions, setConditions }
}

export default useConditions
