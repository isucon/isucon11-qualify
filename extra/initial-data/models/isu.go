package models

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/isucon/isucon11-qualify/bench/random"
)

type Isu struct {
	User         User
	JIAIsuUUID   string
	Name         string
	Image        []byte
	JIACatalogID string
	Character    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewIsu(user User) Isu {
	u, _ := uuid.NewRandom()
	createdAt := random.TimeAfterArg(user.CreatedAt)

	image := defaultImage()

	return Isu{
		user,
		u.String(),
		random.IsuName(),
		image,
		random.CatalogID(),
		random.Character(),
		createdAt,
		createdAt,
	}
}

func NewIsuWithCreatedAt(user User, createdAt time.Time) Isu {
	u, _ := uuid.NewRandom()
	image := defaultImage()

	return Isu{
		user,
		u.String(),
		random.IsuName(),
		image,
		random.CatalogID(),
		random.Character(),
		createdAt,
		createdAt,
	}
}

func defaultImage() []byte {
	bytes, err := ioutil.ReadFile(defaultImagePath)
	if err != nil {
		log.Fatalf("%+v", fmt.Errorf("%w", err))
	}
	return bytes
}

func (i Isu) WithUpdateName() Isu {
	i.Name = random.IsuName()
	i.UpdatedAt = random.TimeAfterArg(i.UpdatedAt)
	return i
}
func (i Isu) WithUpdateImage() Isu {
	i.Image = random.Image()
	i.UpdatedAt = random.TimeAfterArg(i.UpdatedAt)
	return i
}

func (i Isu) Create() error {
	if _, err := db.Exec("INSERT INTO isu(`jia_isu_uuid`,`name`,`image`,`character`,`jia_user_id`,`created_at`,`updated_at`) VALUES (?,?,?,?,?,?,?,?)",
		i.JIAIsuUUID, i.Name, i.Image, i.Character, i.User.JIAUserID,
		i.CreatedAt, i.UpdatedAt,
	); err != nil {
		return fmt.Errorf("insert isu: %w", err)
	}
	return nil
}
