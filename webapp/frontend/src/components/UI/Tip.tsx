interface Props {
  variant: 'info' | 'warning' | 'sitting' | 'critical'
}

const Tip = ({ variant }: Props) => {
  const color = (() => {
    switch (variant) {
      case 'info':
        return 'bg-status-info text-status-info'
      case 'warning':
        return 'bg-status-warning text-status-warning'
      case 'sitting':
        return 'bg-status-sitting text-status-sitting'
      case 'critical':
        return 'bg-status-critical text-status-critical'
    }
  })()
  const className = `h-6 rounded-2xl px-4 font-medium text-center ${color}`
  return <div className={className}>{variant}</div>
}

export default Tip
