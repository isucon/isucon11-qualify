import Controls from './Controls'
import logo from '/@/assets/logo.png'

const PageHeader = () => {
  return (
    <header className="flex p-2 items-center bg-primary-400">
      <img src={logo} alt="isucondition" />
      <Controls />
    </header>
  )
}

export default PageHeader
