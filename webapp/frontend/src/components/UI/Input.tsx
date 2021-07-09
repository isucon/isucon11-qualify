import { DetailedHTMLProps, InputHTMLAttributes } from 'react'

interface Props {
  label: string
  value: string
  setValue: (newValue: string) => void
  classname?: string
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
  classname,
  inputProps
}: Props & InputProps) => {
  return (
    <label className={'flex flex-col ' + classname}>
      {label}
      <input
        {...inputProps}
        className="px-2 py-1 bg-teritary border border-solid border-outline rounded"
        value={value}
        onChange={e => setValue(e.target.value)}
      ></input>
    </label>
  )
}

export default Input
