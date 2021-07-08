package model

import "strconv"

func getKey(targetIP string, targetPort int, isuUUID string) string {
	return isuUUID + targetIP + strconv.Itoa(targetPort)
}
