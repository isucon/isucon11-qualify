interface Props {
  start_time?: string
  end_time?: string
  setStartTime: (newTime: string) => void
  setEndTime: (newTime: string) => void
}

const TimeInputs = ({
  start_time,
  end_time,
  setStartTime,
  setEndTime
}: Props) => {
  return (
    <label className="flex flex-col">
      時間指定
      <div className="flex items-center">
        <input
          className="w-50 border-primary px-2 py-1 h-8 border border-solid rounded"
          value={start_time ? start_time : ''}
          onChange={e => setStartTime(e.target.value)}
          placeholder={'2020/09/12 11:22:33'}
        ></input>
        <div className="m-0.5 text-xl">~</div>
        <input
          className="w-50 border-primary px-2 py-1 h-8 border border-solid rounded"
          value={end_time ? end_time : ''}
          onChange={e => setEndTime(e.target.value)}
          placeholder={'2021/08/21 11:22:33'}
        ></input>
      </div>
    </label>
  )
}

export default TimeInputs
