import { DetailedHTMLProps, InputHTMLAttributes } from 'react'

interface Props {
  label: string
  value: string
  setValue: (newValue: string) => void
  classname?: string
}

type InputProps = DetailedHTMLProps<
  InputHTMLAttributes<HTMLInputElement>,
  HTMLInputElement
>

const Input = ({
  label,
  value,
  setValue,
  classname,
  ...inputProps
}: Props & InputProps) => {
  return (
    <label className={'flex flex-col ' + classname}>
      {label}
      <input
        className="p-1 bg-teritary border-2 border-outline rounded"
        value={value}
        onChange={e => setValue(e.target.value)}
        {...inputProps}
      ></input>
    </label>
  )
}

export default Input
