const NowLoading = () => {
  const base =
    'bg-accent-primary opacity-80 w-4 h-35 m-2 rounded-md inline-block'
  return (
    <div>
      <span className={base + ' animate-loader0'} />
      <span className={base + ' animate-loader1'} />
      <span className={base + ' animate-loader2'} />
      <span className={base + ' animate-loader3'} />
      <span className={base + ' animate-loader4'} />
    </div>
  )
}

export default NowLoading
