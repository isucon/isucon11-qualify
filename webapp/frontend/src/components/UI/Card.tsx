interface Props {
  children: JSX.Element
}

const Card = ({ children }: Props) => {
  return (
    <div className="bg-secondary drop-shadow-lg min-w-min px-16 py-12 w-full max-w-4xl h-full border rounded">
      {children}
    </div>
  )
}

export default Card
