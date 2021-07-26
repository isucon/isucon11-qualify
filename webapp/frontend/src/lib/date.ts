export const getNowDate = () => {
  const localDateStr = new Date().toLocaleDateString()
  return new Date(localDateStr)
}

export const dateToTimestamp = (date: Date) => {
  return Math.floor(date.getTime() / 1000)
}

export const timestampToDate = (timestamp: number) => {
  return new Date(timestamp * 1000)
}
