import { Link } from 'react-router-dom'
import Controls from './Controls'
import logo from '/@/assets/logo.png'

const PageHeader = () => {
  return (
    <header className="h-14 flex items-center p-2 bg-primary">
      <Link to="/">
        <img src={logo} alt="isucondition" className="cursor-pointer" />
      </Link>
      <Controls />
    </header>
  )
}

export default PageHeader
