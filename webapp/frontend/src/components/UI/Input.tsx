interface Props {
  label: string
  value: string
  setValue: (newValue: string) => void
  classname?: string
}

const Input = <T extends Props>({
  label,
  value,
  setValue,
  classname,
  ...inputProps
}: T) => {
  return (
    <label className={'flex flex-col ' + classname}>
      {label}
      <input
        className="p-1 bg-teritary border-2 border-outline"
        value={value}
        onChange={e => setValue(e.target.value)}
        {...inputProps}
      ></input>
    </label>
  )
}

export default Input
