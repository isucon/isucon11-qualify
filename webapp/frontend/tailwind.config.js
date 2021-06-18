import colors from 'windicss/colors'

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
