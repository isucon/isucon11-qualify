import Tip from '/@/components/UI/Tip'

const TrendHeadeer = () => {
  return (
    <div className="grid-cols-trend grid p-2">
      <div className="flex justify-center">
        <div>せいかく</div>
      </div>
      <div className="flex gap-4 items-center justify-center">
        <Tip variant="info" label="バッチリ" />
        <Tip variant="warning" label="ぼちぼち" />
        <Tip variant="critical" label="こわれた" />
      </div>
    </div>
  )
}

export default TrendHeadeer
