const parseQuery = (query: string) => {
  const params: { [key: string]: string } = {}
  let isQuote = false
  let wordStack = ''
  let lastlabel = ''
  for (const q of query) {
    switch (q) {
      case `"`: {
        if (!isQuote) {
          isQuote = true
          break
        }
        params[lastlabel] = wordStack
        wordStack = ''
        break
      }
      case `:`: {
        if (isQuote) {
          wordStack += q
        } else {
          lastlabel = wordStack
          wordStack = ''
        }
        break
      }
      case ' ': {
        if (isQuote) {
          wordStack += q
        } else {
          if (lastlabel && wordStack) {
            params[lastlabel] = wordStack
          }
          lastlabel = ''
          wordStack = ''
        }
        break
      }
      default: {
        wordStack += q
        break
      }
    }
  }
  if (lastlabel && wordStack) {
    params[lastlabel] = wordStack
  }
  return params
}

export default parseQuery
