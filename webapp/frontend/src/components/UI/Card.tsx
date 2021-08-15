interface Props {
  children: JSX.Element
}

const Card = ({ children }: Props) => {
  return (
    <div className="bg-secondary px-16 py-12 w-full max-w-4xl border rounded">
      {children}
    </div>
  )
}

export default Card
