interface Props {
  label: string
  classname?: string
}
const Button = ({ label, classname }: Props) => {
  return (
    <button
      className={
        'px-3 py-1 h-8 leading-4 border border-outline rounded focus:outline-none ' +
        classname
      }
    >
      {label}
    </button>
  )
}

export default Button
