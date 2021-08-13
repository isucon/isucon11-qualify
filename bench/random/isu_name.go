package random

import "math/rand"

func IsuName() string {
	return generatePrefix() + generateSuffix()
}

var prefixData = []string{
	"ゲーミング", "オフィス", "リビング", "ダイニング", "キャンピング", "折りたたみ", "ラウンジ", "マッサージ", "カウンター", "木製", "金属製", "手作り", "ISUCON", "スーパー", "ハイパー", "ミラクル", "パイプ", "量産型", "専用", "高級", "はりぼて", "金メッキ", "エルゴノミクス", "シャア専用", "北欧風", "座・",
}
var suffixData = []string{
	"1号", "ベンチ", "ソファ", "カウチ", "チェア", "座椅子", "三脚", "四脚", "一脚", "ポチ", "ミケ", "タマ", "バランスボール", "サイコー", "スツール", "樽", "座布団", "Mk-Ⅱ", "玉座", "ISU",
}

func generatePrefix() string {
	return prefixData[rand.Intn(len(prefixData))]
}

func generateSuffix() string {
	return suffixData[rand.Intn(len(suffixData))]
}
