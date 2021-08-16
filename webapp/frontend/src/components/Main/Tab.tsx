import TabLink from './TabLink'

const Tab = ({ id }: { id: string }) => {
  return (
    <div className="flex">
      <TabLink to={`/isu/${id}`} label="詳細" />
      <TabLink to={`/isu/${id}/condition`} label="状態" />
      <TabLink to={`/isu/${id}/graph`} label="グラフ" />
    </div>
  )
}

export default Tab
