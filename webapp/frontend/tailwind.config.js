// eslint-disable-next-line @typescript-eslint/no-var-requires
const colors = require('windicss/colors')

module.exports = {
  theme: {
    backgroundColor: theme => ({
      ...theme('colors'),
      primary: colors.indigo
    }),
    textColor: theme => ({
      ...theme('colors'),
      primary: colors.gray
    })
  }
}
