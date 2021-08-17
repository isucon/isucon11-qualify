import { Link } from 'react-router-dom'
import Controls from './Controls'
import logo_white from '/@/assets/logo_white.svg'

const PageHeader = () => {
  return (
    <header className="h-18 bg-accent-primary flex items-center pl-6 pr-8">
      <Link to="/">
        <img
          src={logo_white}
          alt="isucondition"
          className="h-11 cursor-pointer"
        />
      </Link>
      <Controls />
    </header>
  )
}

export default PageHeader
