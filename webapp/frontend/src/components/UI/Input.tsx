interface Props {
  label: string
  value: string
  setValue: (newValue: string) => void
}

const Input = <T extends Props>({
  label,
  value,
  setValue,
  ...inputProps
}: T) => {
  return (
    <label>
      {label}
      <input
        className="border-dark-200 border"
        value={value}
        onChange={e => setValue(e.target.value)}
        {...inputProps}
      ></input>
    </label>
  )
}

export default Input
