interface Props {
  children: JSX.Element
}

const Card = ({ children }: Props) => {
  return (
    <div className="max-w-4xl bg-secondary border border-outline rounded">
      {children}
    </div>
  )
}

export default Card
