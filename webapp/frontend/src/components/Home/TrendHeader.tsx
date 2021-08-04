import Tip from '../UI/Tip'

const TrendHeadeer = () => {
  return (
    <div className="grid grid-cols-trend p-2">
      <div className="flex justify-center">
        <div>性格</div>
      </div>
      <div className="flex gap-4 items-center justify-center">
        <div>ISUの数</div>
        <Tip variant="info" />
        <Tip variant="warning" />
        <Tip variant="critical" />
      </div>
    </div>
  )
}

export default TrendHeadeer
