import { ButtonHTMLAttributes, DetailedHTMLProps } from 'react'

interface Props {
  children: JSX.Element
  onClick?: () => void
}

type ButtonProps = DetailedHTMLProps<
  ButtonHTMLAttributes<HTMLButtonElement>,
  HTMLButtonElement
>

const IconButton = ({ children, onClick, ...props }: Props & ButtonProps) => {
  return (
    <button
      className="flex items-center focus:outline-none disabled:opacity-25 disabled:cursor-not-allowed"
      onClick={onClick}
      {...props}
    >
      {children}
    </button>
  )
}

export default IconButton
