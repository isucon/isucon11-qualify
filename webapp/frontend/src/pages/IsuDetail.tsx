import { useEffect } from 'react'
import { useState } from 'react'
import { useParams } from 'react-router-dom'
import apis, { Isu } from '../lib/apis'

const IsuDetail = () => {
  const [isu, setIsu] = useState<Isu | null>(null)
  const { id } = useParams<{ id: string }>()

  useEffect(() => {
    const load = async () => {
      setIsu(await apis.getIsu(id))
    }
    load()
  }, [id])

  if (!isu) {
    return <div>loading</div>
  }
  return (
    <div>
      <div>椅子詳細</div>
      <div>{isu.name}</div>
    </div>
  )
}

export default IsuDetail
