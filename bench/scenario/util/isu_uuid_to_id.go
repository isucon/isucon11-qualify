package util

import "fmt"

var (
	// set は Isu の new 時にしか行われず append only であるため、特にセマフォは作らない
	jiaIsuUUID2IsuID map[string]int
)

func init() {
	jiaIsuUUID2IsuID = make(map[string]int, 2048) // TODO 適当なので考える
}

func SetKeyValue(jiaIsuUUID string, isuID int) {
	jiaIsuUUID2IsuID[jiaIsuUUID] = isuID
}

func GetIsuIDFromJIAIsuUUID(jiaIsuUUID string) (int, error) {
	isuID, ok := jiaIsuUUID2IsuID[jiaIsuUUID]
	if !ok {
		return 0, fmt.Errorf("jia_isu_uuid が正しくありません")
	}
	return isuID, nil
}
