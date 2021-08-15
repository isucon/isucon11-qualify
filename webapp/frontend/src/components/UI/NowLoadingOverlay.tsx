import NowLoading from './NowLoading'

const NowLoadingOverlay = () => {
  return (
    <div className="absolute top-0 flex items-center justify-center w-full h-full bg-white bg-opacity-40">
      <NowLoading />
    </div>
  )
}

export default NowLoadingOverlay
