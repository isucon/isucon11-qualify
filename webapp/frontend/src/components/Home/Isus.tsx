import { useEffect } from 'react'
import { useState } from 'react'
import apis, { GetIsuListResponse } from '/@/lib/apis'
import IsuList from '/@/components/UI/IsuList'

const LIMIT = 4

const Isus = () => {
  const [isus, setIsus] = useState<GetIsuListResponse[]>([])
  useEffect(() => {
    const fetchIsus = async () => {
      setIsus(await apis.getIsus({ limit: LIMIT }))
    }
    fetchIsus()
  }, [setIsus])

  return (
    <div>
      <h2 className="mb-6 text-xl font-bold">ISU</h2>
      <IsuList isus={isus} />
      <div className="flex gap-12 items-center justify-center mt-8">
        {/* TODO: ISU一覧へのリンク
        <Link to="/search">
          <Button label="もっと見る" />
        </Link> */}
      </div>
    </div>
  )
}

export default Isus
