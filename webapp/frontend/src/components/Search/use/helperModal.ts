import { RefObject, useCallback, useEffect, useState } from 'react'

const useHelperModal = (inputRef: RefObject<HTMLInputElement>) => {
  const [rect, setRect] = useState({ x: 0, y: 0, width: 0 })
  const getInputRect = useCallback(() => {
    if (inputRef.current) {
      const rect = inputRef.current.getBoundingClientRect()
      const offset = { x: window.pageXOffset, y: window.pageYOffset }
      return {
        x: rect.x + offset.x,
        y: rect.bottom + offset.y,
        width: rect.width
      }
    }
  }, [inputRef])

  const [isOpen, setIsOpen] = useState(false)
  const toggle = useCallback(() => setIsOpen(!isOpen), [isOpen])

  useEffect(() => {
    const newRect = getInputRect()
    if (newRect) {
      setRect(newRect)
    }
  }, [isOpen, getInputRect])

  const [resizeObserver, setResizeObserver] = useState<ResizeObserver | null>(
    null
  )
  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.addEventListener('click', () => {
        toggle()
      })
      if (!resizeObserver) {
        const ro = new ResizeObserver(() => {
          if (inputRef.current) {
            const newRect = getInputRect()
            if (newRect) {
              setRect(newRect)
            }
          }
        })
        ro.observe(inputRef.current)
        setResizeObserver(ro)
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [inputRef, setRect, getInputRect, setResizeObserver])
  return { isOpen, toggle, rect }
}

export default useHelperModal
