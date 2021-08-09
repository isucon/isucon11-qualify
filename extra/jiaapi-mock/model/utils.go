package model

func getKey(targetURL string, isuUUID string) string {
	return isuUUID + targetURL
}
