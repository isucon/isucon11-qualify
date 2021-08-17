interface Props {
  top?: boolean
}

const NowLoading = ({ top }: Props) => {
  const base =
    'bg-accent-primary opacity-80 w-4 h-35 m-2 rounded-md inline-block'
  return (
    <div
      className={
        'bg-opacity-40 absolute top-0 flex justify-center w-full h-full bg-white ' +
        (top ? 'pt-12' : 'items-center')
      }
    >
      <span>
        <span className={base + ' animate-loader0'} />
        <span className={base + ' animate-loader1'} />
        <span className={base + ' animate-loader2'} />
        <span className={base + ' animate-loader3'} />
        <span className={base + ' animate-loader4'} />
      </span>
    </div>
  )
}

export default NowLoading
