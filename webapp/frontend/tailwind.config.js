import colors from 'windicss/colors'

module.exports = {
  theme: {
    backgroundColor: theme => ({
      ...theme('colors'),
      primary: colors.indigo
    })
  }
}
