import { DetailedHTMLProps, ButtonHTMLAttributes } from 'react'

interface Props {
  label: string
  children?: JSX.Element
  customClass?: string
}

type ButtonProps = DetailedHTMLProps<
  ButtonHTMLAttributes<HTMLButtonElement>,
  HTMLButtonElement
>

const Button = ({
  label,
  children,
  customClass,
  ...buttonProps
}: Props & ButtonProps) => {
  return (
    <button
      className={
        'flex items-center justify-center focus:outline-none disabled:opacity-40 disabled:cursor-default ' +
        customClass
      }
      {...buttonProps}
    >
      {children ? children : label}
    </button>
  )
}

export default Button
