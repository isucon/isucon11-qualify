import NowLoading from './NowLoading'

const NowLoadingOverlay = () => {
  return (
    <div className="absolute top-0 flex items-center justify-center w-full h-full bg-gray-500 bg-opacity-25">
      <NowLoading />
    </div>
  )
}

export default NowLoadingOverlay
