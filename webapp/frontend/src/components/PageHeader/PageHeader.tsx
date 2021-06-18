import { Link } from 'react-router-dom'
import Controls from './Controls'
import logo from '/@/assets/logo.png'

const PageHeader = () => {
  return (
    <header className="flex p-2 items-center bg-primary-400">
      <Link to="/">
        <img src={logo} alt="isucondition" className="cursor-pointer" />
      </Link>
      <Controls />
    </header>
  )
}

export default PageHeader
