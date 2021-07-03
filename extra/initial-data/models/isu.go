package models

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/isucon/isucon11-qualify/extra/initial-data/random"
)

type Isu struct {
	User         User
	JIAIsuUUID   string
	Name         string
	Image        []byte
	JIACatalogID string
	Character    string
	IsDeleted    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewIsu(user User) Isu {
	u, _ := uuid.NewRandom()
	createdAt := random.Time()

	updatedAt := createdAt
	image := defaultImage()

	return Isu{
		user,
		u.String(),
		random.IsuName(),
		image,
		random.CatalogID(),
		random.Character(),
		false,
		createdAt,
		updatedAt,
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
func (i Isu) WithDelete() Isu {
	i.IsDeleted = true
	i.UpdatedAt = random.TimeAfterArg(i.UpdatedAt)
	return i
}

func (i Isu) Create() error {
	if _, err := db.Exec("INSERT INTO isu VALUES (?,?,?,?,?,?,?,?,?)",
		i.JIAIsuUUID, i.Name, i.Image, i.Character, i.JIACatalogID, i.User.JIAUserID,
		i.IsDeleted, i.CreatedAt, i.UpdatedAt,
	); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}
