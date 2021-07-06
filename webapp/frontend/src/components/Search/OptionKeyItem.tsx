interface Props {
  keyName: string
  description: string
  onClick: () => void
}

const OptionKeyItem = ({ keyName, description, onClick }: Props) => {
  return (
    <div
      onClick={onClick}
      className="grid grid-cols-search px-4 py-1 hover:bg-teritary cursor-pointer"
    >
      <div>{keyName}</div>
      <div>{description}</div>
    </div>
  )
}

export default OptionKeyItem
