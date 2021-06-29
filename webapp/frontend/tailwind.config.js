module.exports = {
  theme: {
    backgroundColor: theme => ({
      ...theme('colors'),
      primary: '#EF724A',
      secondary: '#FFFFFF',
      teritary: '#F1F1F1',
      button: '#21394B',
      'status-info': '#94EFBC',
      'status-warning': '#FFEF5B',
      'status-sitting': '#FFBC7E',
      'status-critical': '#F69898'
    }),
    textColor: theme => ({
      ...theme('colors'),
      primary: '#241E12',
      secondary: '#6B6965',
      'white-primary': '#FFFFFF',
      'white-secondary': '#FFFFFF',
      'accent-primary': '#EF724A'
    }),
    borderColor: theme => ({
      ...theme('colors'),
      outline: '#E1E5E6',
      'accent-primary': '#EF724A'
    })
  }
}
