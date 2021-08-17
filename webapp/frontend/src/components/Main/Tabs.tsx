import { Link, useLocation } from 'react-router-dom'

const Tabs = ({ id }: { id: string }) => {
  return (
    <div className="w-max flex">
      <TabLink to={`/isu/${id}`} label="詳細" />
      <TabLink to={`/isu/${id}/condition`} label="状態" />
      <TabLink to={`/isu/${id}/graph`} label="グラフ" />
    </div>
  )
}

interface Props {
  to: string
  label: string
}

const TabLink = ({ to, label }: Props) => {
  const { pathname } = useLocation()

  const isSelected = pathname === to
  return (
    <div
      className={`w-20  pb-1 flex justify-center ${
        isSelected
          ? 'border-b-2 text-accent-primary border-accent-primary font-bold'
          : 'text-secondary'
      }`}
    >
      <Link to={to}>{label}</Link>
    </div>
  )
}

export default Tabs
