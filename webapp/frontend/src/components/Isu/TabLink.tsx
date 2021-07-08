import { Link, useLocation } from 'react-router-dom'

interface Props {
  to: string
  label: string
}
const TabLink = ({ to, label }: Props) => {
  const { pathname } = useLocation()

  // TODO: これだとtrailing slashが判別できないからすべきか含めて考える
  const isSelected = pathname === to
  return (
    <div
      className={`w-16 flex justify-center ${
        isSelected
          ? 'border-b-2 text-accent-primary border-accent-primary font-bold'
          : 'text-secondary'
      }`}
    >
      <Link to={to}>{label}</Link>
    </div>
  )
}

export default TabLink
