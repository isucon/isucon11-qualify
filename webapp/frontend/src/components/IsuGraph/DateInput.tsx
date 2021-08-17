import { useState, useEffect } from 'react'

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
    <input
      className="border-primary focus:border-primary w-30 px-2 py-1 h-8 text-center border border-solid rounded focus:outline-none shadow-none"
      value={tmpDay}
      onChange={e => setTmpDay(e.target.value)}
      onKeyPress={e => {
        if (e.key === 'Enter') {
          e.preventDefault()
          setDay(tmpDay)
        }
      }}
    />
  )
}

export default DateInput
