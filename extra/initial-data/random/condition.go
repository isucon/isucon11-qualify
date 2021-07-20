package random

import "math/rand"

func Condition() (isSitting bool, isDirty bool, isOverweight bool, isBroken bool) {
	if rand.Intn(3) < 1 { // 1/3
		isSitting = true
	}
	if rand.Intn(4) < 1 { // 1/4
		isDirty = true
	}
	if rand.Intn(4) < 1 { // 1/4
		isOverweight = true
	}
	if rand.Intn(4) < 1 { // 1/4
		isBroken = true
	}
	return
}

func IsSittingFromLastCondition(wasSit bool) (isSitting bool) {
	if wasSit {
		if rand.Intn(6) < 5 { // 5/6
			isSitting = true
		}
	} else {
		if rand.Intn(3) < 1 { // 1/3
			isSitting = true
		}
	}
	return
}

func IsDirtyFromLastCondition(wasDirty bool) (isDirty bool) {
	if wasDirty {
		if rand.Intn(2) < 1 { // 1/2
			isDirty = true
		}
	} else {
		if rand.Intn(4) < 1 { // 1/3
			isDirty = true
		}
	}
	return
}

func IsOverweightFromLastCondition(wasOverweight bool) (isOverweight bool) {
	if wasOverweight {
		if rand.Intn(2) < 1 { // 1/2
			isOverweight = true
		}
	} else {
		if rand.Intn(4) < 1 { // 1/3
			isOverweight = true
		}
	}
	return
}

func IsBrokenFromLastCondition(wasBroken bool) (isBroken bool) {
	if wasBroken {
		if rand.Intn(2) < 1 { // 1/2
			isBroken = true
		}
	} else {
		if rand.Intn(4) < 1 { // 1/3
			isBroken = true
		}
	}
	return
}
