import { useEffect } from 'react'
import { useState } from 'react'
import Button from '../UI/Button'

interface Props {
  day: string
  fetchGraphs: (payload: { day: string }) => Promise<void>
}

const DateInput = ({ day, fetchGraphs }: Props) => {
  const [tmpDay, setTmpDay] = useState(day)

  return (
    <div>
      <input value={tmpDay} onChange={e => setTmpDay(e.target.value)}></input>
      <Button
        label="検索"
        onClick={() => {
          fetchGraphs({ day: tmpDay })
        }}
      />
    </div>
  )
}

const dateToStr = (date: Date) => {
  return `${date.getUTCFullYear()}/${pad0(date.getUTCMonth() + 1)}/${pad0(
    date.getUTCDate()
  )} `
}
const pad0 = (num: number) => ('0' + num).slice(-2)

export default DateInput
