interface Props {
  variant: 'info' | 'warning' | 'sitting' | 'critical'
}

const Tip = ({ variant }: Props) => {
  const color = (() => {
    switch (variant) {
      case 'info':
        return 'bg-status-info'
      case 'warning':
        return 'bg-status-warning'
      case 'sitting':
        return 'bg-status-sitting'
      case 'critical':
        return 'bg-status-critical'
    }
  })()
  const className = `h-8 rounded-2xl py-1 px-4 font-medium text-center ${color}`
  return <div className={className}>{variant}</div>
}

export default Tip
