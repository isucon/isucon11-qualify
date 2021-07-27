export const getNowDate = () => {
  return new Date()
}

export const getTodayDate = () => {
  const date = getNowDate()
  date.setHours(0, 0, 0, 0)
  return date
}

export const dateToTimestamp = (date: Date) => {
  return Math.floor(date.getTime() / 1000)
}

export const timestampToDate = (timestamp: number) => {
  return new Date(timestamp * 1000)
}
