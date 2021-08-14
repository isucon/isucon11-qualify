import { useState, useEffect } from 'react'
import AutosizeInput from 'react-input-autosize'
import Button from '../UI/Button'
import Input from '../UI/Input'

interface Props {
  day: string
  fetchGraphs: (payload: { day: string }) => Promise<void>
}

const DateInput = ({ day, fetchGraphs }: Props) => {
  const [tmpDay, setTmpDay] = useState(day)

  useEffect(() => {
    setTmpDay(day)
  }, [day, setTmpDay])

  return (
    <div className="flex gap-8 w-full">
      <AutosizeInput
        value={tmpDay}
        onChange={e => setTmpDay(e.target.value)}
        onKeyPress={e => {
          if (e.key === 'Enter') {
            e.preventDefault()
            fetchGraphs({ day: tmpDay })
          }
        }}
      />
    </div>
  )
}

export default DateInput
