package random

func MessageWithCondition(isSitting, isDirty, isOverweight, isBroken bool, charactor string) string {
	var score int
	if isSitting {
		score += 1
	}
	if isDirty {
		score += 2
	}
	if isOverweight {
		score += 4
	}
	if isBroken {
		score += 8
	}
	switch score {
	case 1:
		// TODO
	}
	return "TODO"
}
