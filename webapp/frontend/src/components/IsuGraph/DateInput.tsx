import { useState, useEffect } from 'react'
import AutosizeInput from 'react-input-autosize'

interface Props {
  day: string
  setDay: (day: string) => Promise<void>
}

const DateInput = ({ day, setDay }: Props) => {
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
            setDay(tmpDay)
          }
        }}
      />
    </div>
  )
}

export default DateInput
