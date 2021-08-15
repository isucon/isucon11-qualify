import { MouseEventHandler } from 'react'
import Button from './Button'

interface Props {
  label: string
  onClick?: MouseEventHandler<HTMLButtonElement>
  disabled?: boolean
}

const ButtonSub = ({ label, onClick, disabled }: Props) => {
  return (
    <Button
      label={label}
      customClass="px-3 py-1 h-8 leading-4 border rounded"
      onClick={onClick}
      disabled={disabled}
    />
  )
}

export default ButtonSub
