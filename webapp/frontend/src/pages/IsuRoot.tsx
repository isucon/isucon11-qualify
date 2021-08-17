import { useEffect, useState } from 'react'
import { Redirect, Switch, useParams } from 'react-router-dom'
import apis, { Isu } from '/@/lib/apis'
import GuardedRoute from '/@/router/GuardedRoute'
import SubHeader from '/@/components/Main/SubHeader'
import Card from '/@/components/UI/Card'
import IsuCondition from './IsuCondition'
import IsuDetail from './IsuDetail'
import IsuGraph from './IsuGraph'
import NowLoading from '/@/components/UI/NowLoading'
import toast from 'react-hot-toast'

const IsuRoot = () => {
  const [isu, setIsu] = useState<Isu | null>(null)
  const { id } = useParams<{ id: string }>()
  const [notFound, setNotFound] = useState(false)

  useEffect(() => {
    const load = async () => {
      try {
        setIsu(await apis.getIsu(id))
      } catch (e) {
        if (e.response.status === 404) {
          setNotFound(true)
        }
        toast.error(e.response.data)
      }
    }
    load()
  }, [id, setNotFound])

  if (!isu) {
    if (notFound) {
      return <Redirect to={`/`} />
    }
    return <NowLoading />
  }

  return (
    <div>
      <SubHeader isu={isu} />
      <div className="flex flex-col gap-10 items-center p-10">
        <Card>
          <DefineRoutes isu={isu} />
        </Card>
      </div>
    </div>
  )
}

const DefineRoutes = ({ isu }: { isu: Isu }) => (
  <Switch>
    <GuardedRoute path="/isu/:id" exact>
      <IsuDetail isu={isu} />
    </GuardedRoute>
    <GuardedRoute path="/isu/:id/condition" exact>
      <IsuCondition isu={isu} />
    </GuardedRoute>
    <GuardedRoute path="/isu/:id/graph" exact>
      <IsuGraph isu={isu} />
    </GuardedRoute>
  </Switch>
)

export default IsuRoot
