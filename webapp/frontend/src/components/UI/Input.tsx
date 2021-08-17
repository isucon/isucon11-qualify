import { DetailedHTMLProps, InputHTMLAttributes } from 'react'

interface Props {
  label: string
  value: string
  setValue: (newValue: string) => void
  customClass?: string
  inputProps?: InputProps
}

type InputProps = DetailedHTMLProps<
  InputHTMLAttributes<HTMLInputElement>,
  HTMLInputElement
>

const Input = ({
  label,
  value,
  setValue,
  customClass,
  inputProps
}: Props & InputProps) => {
  return (
    <label className={'flex flex-col ' + customClass}>
      {label}
      <input
        type="text"
        {...inputProps}
        className="border-primary focus:border-primary bg-secondary px-2 py-1 h-8 border-solid rounded focus:outline-none shadow-none"
        value={value}
        onChange={e => setValue(e.target.value)}
      ></input>
    </label>
  )
}

export default Input
