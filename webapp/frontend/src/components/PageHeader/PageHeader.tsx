import { Link } from 'react-router-dom'
import Controls from './Controls'
import logo_white from '/@/assets/logo_white.svg'

const PageHeader = () => {
  return (
    <header className="h-14 bg-accent-primary flex items-center p-2">
      <Link to="/">
        <img
          src={logo_white}
          alt="isucondition"
          className="w-50 ml-2 cursor-pointer"
        />
      </Link>
      <Controls />
    </header>
  )
}

export default PageHeader
