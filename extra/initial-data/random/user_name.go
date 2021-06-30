package random

import (
	"github.com/docker/docker/pkg/namesgenerator"
)

func UserName() string {
	return namesgenerator.GetRandomName(0)
}
