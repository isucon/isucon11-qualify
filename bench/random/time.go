package random

import (
	"math/rand"
	"time"
)

var (
	jst      = time.FixedZone("Asia/Tokyo", 9*60*60)
	baseTime = time.Date(2021, 8, 21, 0, 0, 0, 0, jst)
)

func Time() time.Time {
	subFrom := baseTime.Unix()
	subValue := rand.Int63n(60 * 60 * 24 * 365 / 2) // 0 ~ 半年
	return time.Unix(subFrom-subValue, 0)
}

func TimeAfterArg(t time.Time) time.Time {
	createdAtUnix := t.Unix()
	baseTimeUnix := baseTime.Unix()
	return time.Unix(createdAtUnix+rand.Int63n(baseTimeUnix-createdAtUnix), 0)
}
