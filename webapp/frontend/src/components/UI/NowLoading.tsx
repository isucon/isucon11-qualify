const NowLoading = () => {
  const base = "bg-accent-primary opacity-80 w-4 h-35 m-2 rounded-md inline-block"
  return (
    <div className="flex h-100 justify-center items-center">
      <span>
        <span className={base+" animate-loader0"} />
        <span className={base+" animate-loader1"} />
        <span className={base+" animate-loader2"} />
        <span className={base+" animate-loader3"} />
        <span className={base+" animate-loader4"} />
      </span>
    </div>
  );
}

export default NowLoading
