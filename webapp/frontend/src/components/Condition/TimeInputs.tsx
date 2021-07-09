interface Props {
  times: string[]
  setTimes: (newTimes: string[]) => void
}

const TimeInputs = ({ times, setTimes }: Props) => {
  return (
    <label className="flex flex-col">
      時間指定
      <div className="flex items-center">
        <input
          className="px-2 py-1 w-40 bg-teritary border-2 border-outline rounded"
          value={times[0]}
          onChange={e => setTimes([e.target.value, times[1]])}
          placeholder={'2020/01/01 11:11:11'}
        ></input>
        <div className="text-xl">~</div>
        <input
          className="px-2 py-1 w-40 bg-teritary border-2 border-outline rounded"
          value={times[1]}
          onChange={e => setTimes([times[0], e.target.value])}
          placeholder={'2020/01/01 11:11:11'}
        ></input>
      </div>
    </label>
  )
}

export default TimeInputs
