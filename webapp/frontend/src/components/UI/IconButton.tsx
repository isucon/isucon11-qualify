interface Props {
  children: JSX.Element
  onClick?: () => void
}

const IconButton = ({ children, onClick }: Props) => {
  return (
    <button className="flex items-center focus:outline-none" onClick={onClick}>
      {children}
    </button>
  )
}

export default IconButton
