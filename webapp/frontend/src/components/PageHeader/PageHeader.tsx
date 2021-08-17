import { Link } from 'react-router-dom'
import Controls from './Controls'

const PageHeader = () => {
  return (
    <header className="h-18 bg-accent-primary flex items-center pl-6 pr-8">
      <Link to="/">
        <img
          src="/assets/logo_white.svg"
          alt="isucondition"
          className="h-11 cursor-pointer"
        />
      </Link>
      <Controls />
    </header>
  )
}

export default PageHeader
