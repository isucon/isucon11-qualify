import Tip from '/@/components/UI/Tip'

const TrendHeadeer = () => {
  return (
    <div className="grid-cols-trend grid p-2">
      <div className="flex flex-col justify-center">
        <div className="font-bold">せいかく</div>
        <div className="text-secondary">Last Updated</div>
      </div>
      <div className="flex gap-8 items-center justify-center">
        <Tip variant="info" />
        <Tip variant="warning" />
        <Tip variant="critical" />
      </div>
    </div>
  )
}

export default TrendHeadeer
