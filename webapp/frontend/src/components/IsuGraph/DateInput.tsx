import { useEffect } from 'react'
import { useState } from 'react'
import Button from '../UI/Button'

interface Props {
  date: Date
  search: (date: Date) => Promise<void>
}

const DateInput = ({ date, search }: Props) => {
  const [tmpDate, setTmpDate] = useState(date)

  return (
    <div>
      <input
        value={dateToStr(tmpDate)}
        onChange={e => setTmpDate(new Date(e.target.value))}
      ></input>
      <Button
        label="検索"
        onClick={() => {
          search(tmpDate)
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
