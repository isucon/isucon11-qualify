import { DetailedHTMLProps, ButtonHTMLAttributes } from "react"

interface Props {
  label: string
  classname?: string
}

type ButtonProps = DetailedHTMLProps<ButtonHTMLAttributes<HTMLButtonElement>, HTMLButtonElement>

const Button = ({ label, classname, ...buttonProps }: Props & ButtonProps) => {
  return (
    <button
      className={
        'px-3 py-1 h-8 leading-4 border border-outline rounded focus:outline-none ' +
        classname
      }
      {...buttonProps}
    >
      {label}
    </button>
  )
}

export default Button
