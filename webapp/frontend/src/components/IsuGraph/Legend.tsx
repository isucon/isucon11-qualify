import colors from 'windicss/colors'

const Legend = () => {
  return (
    <div className="flex gap-4">
      <div className="flex gap-1">
        <div className="bg-main mt-1 w-4 h-4"></div>
        <div>sittingの割合</div>
      </div>
      <div className="flex gap-1">
        <div
          className="mt-1 w-4 h-4"
          style={{ backgroundColor: colors.blue[500] }}
        ></div>
        <div>score</div>
      </div>
    </div>
  )
}

export default Legend
