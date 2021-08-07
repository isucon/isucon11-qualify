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

export const getConditionTime = (date: Date) => {
  // 2020/01/01 01:01:01
  return `${date.getFullYear()}/${pad0(date.getMonth() + 1)}/${pad0(
    date.getDate()
  )} ${pad0(date.getHours())}:${pad0(date.getMinutes())}:${pad0(
    date.getSeconds()
  )}`
}

const pad0 = (num: number) => ('0' + num).slice(-2)

export const getPrevDate = (date: Date) => {
  return new Date(date.setDate(date.getDate() - 1))
}

export const getNextDate = (date: Date) => {
  return new Date(date.setDate(date.getDate() + 1))
}
